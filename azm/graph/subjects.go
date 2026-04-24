package graph

import (
	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/aserto-dev/topaz/api/directory/pkg/derr"
	dsc "github.com/aserto-dev/topaz/api/directory/v4"
	dsr "github.com/aserto-dev/topaz/api/directory/v4/reader"

	"github.com/aserto-dev/topaz/azm/mempool"
	"github.com/aserto-dev/topaz/azm/model"
)

var errCycle = errors.New("cycle detected")

type SubjectSearch struct {
	graphSearch
}

func NewSubjectSearch(
	m *model.Model,
	req *dsr.GraphRequest,
	reader RelationReader,
	pool *mempool.RelationsPool,
) (*SubjectSearch, error) {
	params := searchParams(req)
	if err := validate(m, params); err != nil {
		return nil, err
	}

	return &SubjectSearch{graphSearch{
		m:       m,
		params:  params,
		getRels: reader,
		memo:    newSearchMemo(req.GetTrace()),
		explain: req.GetExplain(),
		pool:    pool,
	}}, nil
}

func (s *SubjectSearch) Search() (*dsr.GraphResponse, error) {
	resp := &dsr.GraphResponse{}

	res, err := s.search(s.params)
	if err != nil {
		return resp, err
	}

	resp.Results = res.Subjects()

	if s.explain {
		resp.Explanation, _ = res.Explain()
	}

	resp.Trace = s.memo.Trace()

	return resp, nil
}

func (s *SubjectSearch) search(params *relation) (searchResults, error) {
	status := s.memo.MarkVisited(params)
	switch status {
	case searchStatusComplete:
		return s.memo.visited[*params], nil
	case searchStatusPending:
		// We have a cycle.
		return nil, errCycle
	case searchStatusNew:
	}

	var (
		results searchResults
		err     error
	)

	o := s.m.Objects[params.ot]
	if o.HasRelation(params.rel) {
		results, err = s.searchRelation(params)
	} else {
		results, err = s.searchPermission(params)
	}

	s.memo.MarkComplete(params, results)

	return results, err
}

func (s *SubjectSearch) searchRelation(params *relation) (searchResults, error) {
	r := s.m.Objects[params.ot].Relations[params.rel]

	subjectTypes := []model.ObjectName{}
	if params.tail == "" {
		subjectTypes = append(subjectTypes, params.st)
	}

	steps := s.m.StepRelation(r, subjectTypes...)

	results := searchResults{}

	for _, step := range steps {
		var (
			res searchResults
			err error
		)

		switch {
		case step.IsDirect(), step.IsWildcard():
			res, err = s.findNeighbor(step, params)
		case step.IsSubject():
			res, err = s.searchSubjectRelation(step, params)
		}

		if err != nil {
			return results, err
		}

		results = lo.Assign(results, res)
	}

	return results, nil
}

func (s *SubjectSearch) findNeighbor(step *model.RelationRef, params *relation) (searchResults, error) {
	sid := params.sid.String()
	if step.IsWildcard() {
		sid = model.WildcardSymbol
	}

	req := &dsc.RelationIdentifier{
		ObjectType:  params.ot.String(),
		ObjectId:    params.oid.String(),
		Relation:    params.rel.String(),
		SubjectType: step.Object.String(),
		SubjectId:   sid,
	}

	results := searchResults{}

	relsPtr := s.pool.GetSlice()
	if err := s.getRels(req, s.pool, relsPtr); err != nil {
		return results, err
	}

	for _, rel := range *relsPtr {
		edge := relation{
			ot:  model.ObjectName(rel.GetObjectType()),
			oid: ObjectID(rel.GetObjectId()),
			rel: model.RelationName(rel.GetRelation()),
			st:  model.ObjectName(rel.GetSubjectType()),
			sid: ObjectID(rel.GetSubjectId()),
		}

		var matches searchResults

		if params.tail != "" {
			search := &relation{
				ot:  model.ObjectName(rel.GetSubjectType()),
				oid: ObjectID(rel.GetSubjectId()),
				rel: params.tail,
				st:  params.st,
			}

			res, err := s.search(search)

			switch {
			case err == nil:
				matches = res
			case !errors.Is(err, errCycle):
				return results, err
			}
		} else {
			subj := edge.subject()
			matches = searchResults{*subj: nil}
		}

		for leaf, path := range matches {
			if s.explain {
				path = append(path, searchPath{&edge})
			}

			results[leaf] = path
		}
	}

	s.pool.PutSlice(relsPtr)

	return results, nil
}

func (s *SubjectSearch) searchSubjectRelation(step *model.RelationRef, params *relation) (searchResults, error) {
	results := searchResults{}

	req := &dsc.RelationIdentifier{
		ObjectType:      params.ot.String(),
		ObjectId:        params.oid.String(),
		Relation:        params.rel.String(),
		SubjectType:     step.Object.String(),
		SubjectRelation: step.Relation.String(),
	}

	relsPtr := s.pool.GetSlice()
	if err := s.getRels(req, s.pool, relsPtr); err != nil {
		return results, err
	}
	defer s.pool.PutSlice(relsPtr)

	for _, rel := range *relsPtr {
		current := relationFromProto(rel)

		if params.srel == model.RelationName(rel.GetSubjectRelation()) && params.st == model.ObjectName(rel.GetSubjectType()) {
			// We're searching for a subject relation (not a Check call) and we have a match.
			subj := current.subject()

			var path []searchPath
			if s.explain {
				path = append(results[*subj], searchPath{current})
			}

			results[*subj] = path
		}

		check := &relation{
			ot:   step.Object,
			oid:  ObjectID(rel.GetSubjectId()),
			rel:  step.Relation,
			st:   params.st,
			sid:  params.sid,
			srel: params.srel,
			tail: params.tail,
		}

		res, err := s.search(check)

		switch {
		case errors.Is(err, errCycle):
			continue
		case err != nil:
			return results, err
		}

		if s.explain {
			res = lo.MapValues(res, func(paths []searchPath, _ object) []searchPath {
				return lo.Map(paths, func(p searchPath, _ int) searchPath {
					return append(searchPath{current}, p...)
				})
			})
		}

		results = lo.Assign(results, res)
	}

	return results, nil
}

func (s *SubjectSearch) searchPermission(params *relation) (searchResults, error) {
	o := s.m.Objects[params.ot]

	p := o.Permissions[params.rel]
	if p == nil {
		// This permission isn't defined on the object type.
		return searchResults{}, nil
	}

	results := searchResults{}

	subjTypes := s.possibleSubjects(params)
	if len(lo.Intersect(subjTypes, p.SubjectTypes)) == 0 {
		// The subject type cannot have this permission.
		return results, nil
	}

	terms := p.Terms()
	termChecks := make([][]*relation, 0, len(terms))

	for _, pt := range terms {
		// expand arrow operators.
		expanded, err := s.expandTerm(o, pt, params)
		if err != nil {
			return results, err
		}

		termChecks = append(termChecks, expanded)
	}

	switch {
	case p.IsUnion():
		return s.union(termChecks)
	case p.IsIntersection():
		return s.intersection(termChecks)
	case p.IsExclusion():
		include, err := s.union(termChecks[:1])

		switch {
		case err != nil:
			return results, err
		case include == nil:
			// We have a cycle.
			return nil, errCycle
		case len(include) == 0:
			// Short-circuit: The include term is false, so the permission is false.
			return results, nil
		}

		exclude, err := s.union(termChecks[1:])
		if err != nil {
			return results, err
		}

		return lo.OmitByKeys(include, lo.Keys(exclude)), nil
	}

	return results, derr.ErrUnknown.Msg("unknown permission operator")
}

func (s *SubjectSearch) possibleSubjects(params *relation) []model.ObjectName {
	if params.srel == "" {
		return []model.ObjectName{params.st}
	}

	return s.m.Objects[params.st].Relations[params.srel].SubjectTypes
}

func (s *SubjectSearch) expandTerm(o *model.Object, pt *model.PermissionTerm, params *relation) ([]*relation, error) {
	if pt.IsArrow() {
		if o.HasRelation(pt.Base) {
			return s.expandRelationArrow(pt, params)
		}

		return s.expandPermissionArrow(o, pt, params)
	}

	return []*relation{{ot: params.ot, oid: params.oid, rel: pt.RelOrPerm, st: params.st, sid: params.sid, srel: params.srel}}, nil
}

func (s *SubjectSearch) expandRelationArrow(pt *model.PermissionTerm, params *relation) (relations, error) {
	req := &dsc.RelationIdentifier{
		ObjectType: params.ot.String(),
		ObjectId:   params.oid.String(),
		Relation:   pt.Base.String(),
	}

	relsPtr := s.pool.GetSlice()

	// Resolve the base of the arrow.
	if err := s.getRels(req, s.pool, relsPtr); err != nil {
		return []*relation{}, err
	}

	expanded := lo.Map(*relsPtr, func(rel *dsc.RelationIdentifier, _ int) *relation {
		return &relation{
			ot:   model.ObjectName(rel.GetSubjectType()),
			oid:  ObjectID(rel.GetSubjectId()),
			rel:  lo.Ternary(rel.GetSubjectRelation() == "", pt.RelOrPerm, model.RelationName(rel.GetSubjectRelation())),
			st:   params.st,
			sid:  params.sid,
			srel: params.srel,

			tail: lo.Ternary(rel.GetSubjectRelation() == "", "", pt.RelOrPerm),
		}
	})

	s.pool.PutSlice(relsPtr)

	return expanded, nil
}

func (s *SubjectSearch) expandPermissionArrow(o *model.Object, pt *model.PermissionTerm, params *relation) ([]*relation, error) {
	expanded := []*relation{}

	pBase := o.Permissions[pt.Base]
	for _, subjType := range pBase.SubjectTypes {
		var subs []model.ObjectName

		oBase := s.m.Objects[subjType]
		if oBase.HasRelation(pt.RelOrPerm) {
			subs = oBase.Relations[pt.RelOrPerm].SubjectTypes
		} else {
			subs = oBase.Permissions[pt.RelOrPerm].SubjectTypes
		}

		if !lo.Contains(subs, params.st) {
			// The subject type cannot have this permission.
			continue
		}

		baseSearch := &relation{
			ot:  params.ot,
			oid: params.oid,
			rel: pt.Base,
			st:  subjType,
		}
		res, err := s.search(baseSearch)

		switch {
		case errors.Is(err, errCycle):
			// We have a cycle.
			// We can't expand the permission arrow until we have the results of the base.
			// We leave the object ID empty to indicate that we need to defer the check.
			expanded = append(expanded, &relation{
				ot:   subjType,
				rel:  pt.RelOrPerm,
				st:   params.st,
				sid:  params.sid,
				srel: params.srel,
			})

			continue
		case err != nil:
			return nil, err
		}

		expanded = append(expanded, lo.Map(lo.Keys(res), func(subj object, _ int) *relation {
			return &relation{
				ot:   subj.Type,
				oid:  subj.ID,
				rel:  pt.RelOrPerm,
				st:   params.st,
				sid:  params.sid,
				srel: params.srel,
			}
		})...)
	}

	return expanded, nil
}

func (s *SubjectSearch) union(checks [][]*relation) (searchResults, error) {
	results := searchResults{}
	status := searchStatusPending
	deferred := []*relation{}

	for _, check := range checks {
		var (
			res searchResults
			err error
		)

		switch len(check) {
		case 0:
			res, err = searchResults{}, nil
		case 1:
			if check[0].oid != "" {
				res, err = s.search(check[0])
			} else {
				deferred = append(deferred, check[0])
			}
		default:
			res, err = s.union(lo.Map(check, func(params *relation, _ int) []*relation {
				return []*relation{params}
			}))
		}

		switch {
		case errors.Is(err, errCycle):
			// We have a cycle.
			continue
		case err != nil:
			return res, err
		}

		results = lo.Assign(results, res)
		status = searchStatusComplete
	}

	if len(deferred) > 0 {
		// We have deferred checks that depend on the results of other checks.
		// Fill in the object IDs and re-run the search.
		checks := lo.Map(deferred, func(params *relation, _ int) []*relation {
			return lo.Map(lo.Keys(results), func(subj object, _ int) *relation {
				return &relation{
					ot:   params.ot,
					oid:  subj.ID,
					rel:  params.rel,
					st:   params.st,
					sid:  params.sid,
					srel: params.srel,
				}
			})
		})

		res, err := s.union(checks)
		if err != nil {
			return nil, err
		}

		results = lo.Assign(results, res)
	}

	// return nil if all checks result in a cycle
	return lo.Ternary(status == searchStatusComplete, results, nil), nil
}

func (s *SubjectSearch) intersection(checks [][]*relation) (searchResults, error) {
	results := []searchResults{}
	status := searchStatusPending

	for _, check := range checks {
		// if the base of an arrow operator resolves to multiple objects (e.g. multiple "parents")
		// then a match on any of them is sufficient.
		result, err := s.union([][]*relation{check})

		switch {
		case err != nil:
			return searchResults{}, err
		case result == nil:
			// We have a cycle.
			continue
		case len(result) == 0:
			return result, nil
		}

		results = append(results, result)
		status = searchStatusComplete
	}

	if status == searchStatusPending {
		// All checks result in a cycle.
		return nil, errCycle
	}

	intersection := lo.Reduce(results, func(agg searchResults, item searchResults, i int) searchResults {
		if i == 0 {
			return item
		}

		for subj, paths := range agg {
			itemPaths, inBoth := item[subj]
			if inBoth {
				// add the paths from the current item to the intersection.
				agg[subj] = append(paths, itemPaths...)
			} else {
				// the subject is not in the intersection.
				delete(agg, subj)
			}
		}

		return agg
	}, searchResults{})

	return intersection, nil
}
