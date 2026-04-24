package graph

import (
	"fmt"

	"github.com/aserto-dev/topaz/api/directory/pkg/derr"
	dsc "github.com/aserto-dev/topaz/api/directory/v4"
	dsr "github.com/aserto-dev/topaz/api/directory/v4/reader"
	"github.com/aserto-dev/topaz/azm/mempool"
	"github.com/aserto-dev/topaz/azm/model"

	"github.com/samber/lo"
)

type Checker struct {
	m       *model.Model
	params  *relation
	getRels RelationReader

	memo *checkMemo
	pool *mempool.RelationsPool
}

func NewCheck(m *model.Model, req *dsr.CheckRequest, reader RelationReader, pool *mempool.RelationsPool) *Checker {
	return &Checker{
		m: m,
		params: &relation{
			ot:  model.ObjectName(req.GetObjectType()),
			oid: ObjectID(req.GetObjectId()),
			rel: model.RelationName(req.GetRelation()),
			st:  model.ObjectName(req.GetSubjectType()),
			sid: ObjectID(req.GetSubjectId()),
		},
		getRels: reader,
		memo:    newCheckMemo(req.GetTrace()),
		pool:    pool,
	}
}

func (c *Checker) Check() (bool, error) {
	o := c.m.Objects[c.params.ot]
	if o == nil {
		return false, derr.ErrObjectTypeNotFound.Msg(c.params.ot.String())
	}

	if !o.HasRelOrPerm(c.params.rel) {
		return false, derr.ErrRelationTypeNotFound.Msg(c.params.rel.String())
	}

	status, err := c.check(c.params)
	if err != nil {
		return false, err
	}

	return status == checkStatusTrue, nil
}

func (c *Checker) Trace() []string {
	return c.memo.Trace()
}

func (c *Checker) Reason() string {
	if len(c.memo.cycles) == 0 {
		return ""
	}

	cycles := fmt.Sprintf("%v", c.memo.cycles)

	return "cycles detected: " + cycles
}

func (c *Checker) check(params *relation) (checkStatus, error) {
	prior := c.memo.MarkVisited(params)
	switch prior {
	case checkStatusPending:
		// We have a cycle.
		return prior, nil
	case checkStatusTrue, checkStatusFalse:
		// We already checked this relation.
		return prior, nil
	case checkStatusNew: // this is the first time we're running this check.
	}

	o := c.m.Objects[params.ot]

	var (
		result checkStatus
		err    error
	)

	if o.HasRelation(params.rel) {
		result, err = c.checkRelation(params)
	} else {
		result, err = c.checkPermission(params)
	}

	c.memo.MarkComplete(params, result)

	return result, err
}

func (c *Checker) checkRelation(params *relation) (checkStatus, error) {
	r := c.m.Objects[params.ot].Relations[params.rel]

	var subjectTypes []model.ObjectName
	if params.tail == "" {
		subjectTypes = append(subjectTypes, params.st)
	}

	steps := c.m.StepRelation(r, subjectTypes...)

	// Reuse the same slice in all steps.
	relsPtr := c.pool.GetSlice()
	defer c.pool.PutSlice(relsPtr)

	for _, step := range steps {
		*relsPtr = (*relsPtr)[:0]

		req := &dsc.RelationIdentifier{
			ObjectType:  params.ot.String(),
			ObjectId:    params.oid.String(),
			Relation:    params.rel.String(),
			SubjectType: step.Object.String(),
		}

		switch {
		case step.IsDirect() && (params.tail == "" || params.tail == params.rel):
			req.SubjectId = params.sid.String()
		case step.IsWildcard():
			req.SubjectId = model.WildcardSymbol
		case step.IsSubject():
			req.SubjectRelation = step.Relation.String()
		}

		if err := c.getRels(req, c.pool, relsPtr); err != nil {
			return checkStatusFalse, err
		}

		if status, err := c.checkRelationStep(params, step, *relsPtr); err != nil || status == checkStatusTrue {
			return status, err
		}
	}

	return checkStatusFalse, nil
}

func (c *Checker) checkRelationStep(params *relation, step *model.RelationRef, rels []*dsc.RelationIdentifier) (checkStatus, error) {
	switch {
	case step.IsDirect():
		for _, rel := range rels {
			if status, err := c.checkDirectRelation(params, rel); err != nil || status == checkStatusTrue {
				return status, err
			}
		}

	case step.IsWildcard():
		if len(rels) > 0 {
			// We have a wildcard match.
			return checkStatusTrue, nil
		}

	case step.IsSubject():
		for _, rel := range rels {
			check := &relation{
				ot:   step.Object,
				oid:  ObjectID(rel.GetSubjectId()),
				rel:  step.Relation,
				st:   params.st,
				sid:  params.sid,
				tail: params.tail,
			}
			if status, err := c.check(check); err != nil || status == checkStatusTrue {
				return status, err
			}
		}
	}

	return checkStatusPending, nil
}

func (c *Checker) checkDirectRelation(params *relation, rel *dsc.RelationIdentifier) (checkStatus, error) {
	if params.tail == "" && rel.GetSubjectId() == params.sid.String() {
		return checkStatusTrue, nil
	}

	if params.tail != "" {
		check := &relation{
			ot:  model.ObjectName(rel.GetSubjectType()),
			oid: ObjectID(rel.GetSubjectId()),
			rel: params.tail,
			st:  params.st,
			sid: params.sid,
		}
		if status, err := c.check(check); err != nil {
			return checkStatusFalse, err
		} else if status == checkStatusTrue {
			return status, nil
		}
	}

	return checkStatusPending, nil
}

func (c *Checker) checkPermission(params *relation) (checkStatus, error) {
	p := c.m.Objects[params.ot].Permissions[params.rel]

	if !lo.Contains(p.SubjectTypes, params.st) {
		// The subject type cannot have this permission.
		return checkStatusFalse, nil
	}

	terms := p.Terms()
	termChecks := make([]relations, 0, len(terms))

	for _, pt := range terms {
		// expand arrow operators.
		expanded, err := c.expandTerm(pt, params)
		if err != nil {
			return checkStatusFalse, err
		}

		termChecks = append(termChecks, expanded)
	}

	switch {
	case p.IsUnion():
		return c.checkAny(termChecks)
	case p.IsIntersection():
		return c.checkAll(termChecks)
	case p.IsExclusion():
		include, err := c.checkAny(termChecks[:1])

		switch {
		case err != nil:
			return checkStatusFalse, err
		case include == checkStatusFalse:
			// Short-circuit: The include term is false, so the permission is false.
			return checkStatusFalse, nil
		}

		exclude, err := c.checkAny(termChecks[1:])
		if err != nil {
			return checkStatusFalse, err
		}

		return lo.Ternary(exclude == checkStatusFalse, checkStatusTrue, checkStatusFalse), nil
	}

	return checkStatusFalse, derr.ErrUnknown.Msg("unknown permission operator")
}

func (c *Checker) expandTerm(pt *model.PermissionTerm, params *relation) (relations, error) {
	if pt.IsArrow() {
		query := &dsc.RelationIdentifier{
			ObjectType: params.ot.String(),
			ObjectId:   params.oid.String(),
			Relation:   pt.Base.String(),
		}

		relsPtr := c.pool.GetSlice()

		// Resolve the base of the arrow.
		if err := c.getRels(query, c.pool, relsPtr); err != nil {
			return relations{}, err
		}

		expanded := lo.Map(*relsPtr, func(rel *dsc.RelationIdentifier, _ int) *relation {
			return &relation{
				ot:  model.ObjectName(rel.GetSubjectType()),
				oid: ObjectID(rel.GetSubjectId()),
				rel: lo.Ternary(rel.GetSubjectRelation() == "", pt.RelOrPerm, model.RelationName(rel.GetSubjectRelation())),
				st:  params.st,
				sid: params.sid,

				tail: lo.Ternary(rel.GetSubjectRelation() == "", "", pt.RelOrPerm),
			}
		})

		c.pool.PutSlice(relsPtr)

		return expanded, nil
	}

	return relations{{ot: params.ot, oid: params.oid, rel: pt.RelOrPerm, st: params.st, sid: params.sid}}, nil
}

func (c *Checker) checkAny(checks []relations) (checkStatus, error) {
	result := checkStatusPending

	for _, check := range checks {
		var (
			status checkStatus
			err    error
		)

		switch len(check) {
		case 0:
			status, err = checkStatusFalse, nil
		case 1:
			status, err = c.check(check[0])
		default:
			status, err = c.checkAny(lo.Map(check, func(params *relation, _ int) relations {
				return relations{params}
			}))
		}

		if err != nil {
			return checkStatusFalse, err
		}

		switch status {
		case checkStatusTrue:
			return status, nil
		case checkStatusFalse:
			result = checkStatusFalse
		case checkStatusNew:
			panic("check can never return checkStatusNew")
		case checkStatusPending:
		}
	}

	return result, nil
}

func (c *Checker) checkAll(checks []relations) (checkStatus, error) {
	result := checkStatusPending

	for _, check := range checks {
		// if the base of an arrow operator resolves to multiple objects (e.g. multiple "parents")
		// then a match on any of them is sufficient.
		status, err := c.checkAny([]relations{check})
		if err != nil {
			return checkStatusFalse, err
		}

		switch status {
		case checkStatusFalse:
			return status, nil
		case checkStatusTrue:
			result = checkStatusTrue
		case checkStatusNew:
			panic("check can never return checkStatusNew")
		case checkStatusPending:
		}
	}

	return result, nil
}
