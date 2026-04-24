package graph

import (
	"strings"

	"github.com/aserto-dev/topaz/api/directory/pkg/derr"
	dsc "github.com/aserto-dev/topaz/api/directory/v4"
	dsr "github.com/aserto-dev/topaz/api/directory/v4/reader"
	"github.com/aserto-dev/topaz/azm/mempool"
	"github.com/aserto-dev/topaz/azm/model"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

type ObjectSearch struct {
	subjectSearch  *SubjectSearch
	wildcardSearch *SubjectSearch
}

func NewObjectSearch(m *model.Model, req *dsr.GraphRequest, reader RelationReader, pool *mempool.RelationsPool) (*ObjectSearch, error) {
	params := searchParams(req)
	if err := validate(m, params); err != nil {
		return nil, err
	}

	im := m.Invert()
	// validate the model but allow arrow operators to have permissions
	// in the base of the arrow, not just relations.
	if err := im.Validate(model.AllowPermissionInArrowBase); err != nil {
		log.Err(err).Interface("req", req).Msg("inverted model is invalid")
		// NOTE: we should persist the inverted model instead of computing it on the fly.
		return nil, derr.ErrUnknown.Msg("internal error: unable to search objects.")
	}

	iParams := invertGraphRequest(im, req)

	return &ObjectSearch{
		subjectSearch: &SubjectSearch{graphSearch{
			m:       im,
			params:  iParams,
			getRels: invertedRelationReader(im, reader),
			memo:    newSearchMemo(req.GetTrace()),
			explain: req.GetExplain(),
			pool:    pool,
		}},
		wildcardSearch: &SubjectSearch{graphSearch{
			m:       im,
			params:  wildcardParams(iParams),
			getRels: invertedRelationReader(im, reader),
			memo:    newSearchMemo(req.GetTrace()),
			explain: req.GetExplain(),
			pool:    pool,
		}},
	}, nil
}

func (s *ObjectSearch) Search() (*dsr.GraphResponse, error) {
	resp := &dsr.GraphResponse{}

	res, err := s.subjectSearch.search(s.subjectSearch.params)
	if err != nil {
		return resp, err
	}

	wildRes, err := s.wildcardSearch.search(s.wildcardSearch.params)
	if err != nil {
		return resp, err
	}

	for obj, paths := range wildRes {
		res[obj] = append(res[obj], paths...)
	}

	res = invertResults(res)

	m := s.subjectSearch.m

	memo := s.subjectSearch.memo
	memo.history = append(memo.history, s.wildcardSearch.memo.history...)
	memo.history = lo.Map(memo.history, func(c *searchCall, _ int) *searchCall {
		return &searchCall{
			relation: uninvertRelation(m, c.relation),
			status:   c.status,
		}
	})

	resp.Results = res.Subjects()

	if s.subjectSearch.explain {
		resp.Explanation, _ = res.Explain()
	}

	resp.Trace = memo.Trace()

	return resp, nil
}

func invertGraphRequest(im *model.Model, req *dsr.GraphRequest) *relation {
	rel := model.InverseRelation(
		model.ObjectName(req.GetObjectType()),
		model.RelationName(req.GetRelation()),
		model.RelationName(req.GetSubjectRelation()),
	)
	relPerm := model.PermForRel(rel)

	if im.Objects[model.ObjectName(req.GetSubjectType())].HasPermission(relPerm) {
		rel = relPerm
	} else if req.GetSubjectRelation() != "" {
		rel = model.InverseRelation(
			model.ObjectName(req.GetObjectType()),
			model.RelationName(req.GetRelation()),
			model.RelationName(req.GetSubjectRelation()),
		)
	}

	iReq := &relation{
		ot:  model.ObjectName(req.GetSubjectType()),
		oid: ObjectID(req.GetSubjectId()),
		rel: rel,
		st:  model.ObjectName(req.GetObjectType()),
		sid: ObjectID(req.GetObjectId()),
	}

	o := im.Objects[iReq.ot]
	srPerm := model.GeneratedPermissionPrefix + iReq.rel

	if o.HasRelation(iReq.rel) && o.HasPermission(srPerm) {
		iReq.rel = srPerm
	}

	return iReq
}

func wildcardParams(params *relation) *relation {
	wildcard := *params
	wildcard.oid = model.WildcardSymbol

	return &wildcard
}

func invertedRelationReader(m *model.Model, reader RelationReader) RelationReader {
	return func(r *dsc.RelationIdentifier, relPool MessagePool[*dsc.RelationIdentifier], out *Relations) error {
		ir := uninvertRelation(m, relationFromProto(r))
		if err := reader(ir.asProto(), relPool, out); err != nil {
			return err
		}

		res := *out
		for i, r := range res {
			res[i] = &dsc.RelationIdentifier{
				ObjectType:  r.GetSubjectType(),
				ObjectId:    r.GetSubjectId(),
				Relation:    r.GetRelation(),
				SubjectType: r.GetObjectType(),
				SubjectId:   r.GetObjectId(),
			}
		}

		return nil
	}
}

func uninvertRelation(m *model.Model, r *relation) *relation {
	obj, objRel, _ := strings.Cut(r.rel.String(), model.ObjectNameSeparator)

	rel, srel, found := strings.Cut(objRel, model.SubjectRelationSeparator)
	if found && srel == model.WildcardSymbol {
		srel = ""
	}

	perm := model.PermForRel(model.RelationName(rel))
	if m.Objects[model.ObjectName(obj)].HasPermission(perm) {
		rel = perm.String()
	}

	return &relation{
		ot:   r.st,
		oid:  r.sid,
		rel:  model.RelationName(rel),
		st:   r.ot,
		sid:  r.oid,
		srel: model.RelationName(srel),
	}
}

func invertResults(res searchResults) searchResults {
	return lo.MapValues(res, func(paths []searchPath, obj object) []searchPath {
		return lo.Map(paths, func(p searchPath, _ int) searchPath {
			return lo.Map(p, func(r *relation, _ int) *relation {
				return &relation{
					ot:   r.st,
					oid:  r.sid,
					rel:  r.rel,
					st:   r.ot,
					sid:  r.oid,
					srel: r.srel,
				}
			})
		})
	})
}
