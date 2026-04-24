package graph

import (
	"fmt"
	"strings"

	"github.com/aserto-dev/topaz/api/directory/pkg/derr"
	dsc "github.com/aserto-dev/topaz/api/directory/v4"
	dsr "github.com/aserto-dev/topaz/api/directory/v4/reader"
	"github.com/aserto-dev/topaz/azm/mempool"
	"github.com/aserto-dev/topaz/azm/model"

	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/structpb"
)

type (
	ObjectID  = model.ObjectID
	Relations = []*dsc.RelationIdentifier

	searchPath relations

	object struct {
		Type model.ObjectName
		ID   ObjectID
	}

	// The results of a search is a map where the key is a matching relations
	// and the value is a list of paths that connect the search object and subject.
	searchResults map[object][]searchPath
)

type MessagePool[T any] interface {
	Get() T
	Put(message T)
}

type RelationPool = MessagePool[*dsc.RelationIdentifier]

// RelationReader retrieves relations that match the given filter.
type RelationReader func(*dsc.RelationIdentifier, RelationPool, *Relations) error

// Objects returns the objects from the search results.
func (r searchResults) Objects() []*dsc.ObjectIdentifier {
	return lo.MapToSlice(r, func(o object, _ []searchPath) *dsc.ObjectIdentifier {
		return &dsc.ObjectIdentifier{
			ObjectType: o.Type.String(),
			ObjectId:   o.ID.String(),
		}
	})
}

// Subjects returns the subjects from the search results.
func (r searchResults) Subjects() []*dsc.ObjectIdentifier {
	return lo.MapToSlice(r, func(o object, _ []searchPath) *dsc.ObjectIdentifier {
		return &dsc.ObjectIdentifier{
			ObjectType: o.Type.String(),
			ObjectId:   o.ID.String(),
		}
	})
}

func (r searchResults) Explain() (*structpb.Struct, error) {
	explanation := lo.MapEntries(r, func(obj object, paths []searchPath) (string, any) {
		key := fmt.Sprintf("%s:%s", obj.Type, obj.ID)

		val := lo.Map(paths, func(path searchPath, _ int) any {
			return lo.Map(path, func(rel *relation, _ int) any {
				return rel.String()
			})
		})

		return key, val
	})

	return structpb.NewStruct(explanation)
}

type searchStatus int

const (
	searchStatusNew searchStatus = iota
	searchStatusPending
	searchStatusComplete
)

func (s searchStatus) String() string {
	switch s {
	case searchStatusNew:
		return statusUnknown
	case searchStatusPending:
		return statusPending
	case searchStatusComplete:
		return statusComplete
	default:
		return fmt.Sprintf("invalid: %d", s)
	}
}

type graphSearch struct {
	m       *model.Model
	params  *relation
	getRels RelationReader

	memo    *searchMemo
	explain bool
	pool    *mempool.RelationsPool
}

func validate(m *model.Model, params *relation) error {
	o := m.Objects[params.ot]
	if o == nil {
		return derr.ErrObjectTypeNotFound.Msg(params.ot.String())
	}

	if !o.HasRelOrPerm(params.rel) {
		return derr.ErrRelationTypeNotFound.Msg(params.rel.String())
	}

	if _, ok := m.Objects[params.st]; !ok {
		return derr.ErrObjectTypeNotFound.Msg(params.st.String())
	}

	return nil
}

func searchParams(req *dsr.GraphRequest) *relation {
	return &relation{
		ot:   model.ObjectName(req.GetObjectType()),
		oid:  ObjectID(req.GetObjectId()),
		rel:  model.RelationName(req.GetRelation()),
		st:   model.ObjectName(req.GetSubjectType()),
		sid:  ObjectID(req.GetSubjectId()),
		srel: model.RelationName(req.GetSubjectRelation()),
	}
}

type searchCall struct {
	*relation

	status searchStatus
}

type searchMemo struct {
	visited map[relation]searchResults
	history []*searchCall
}

func newSearchMemo(trace bool) *searchMemo {
	return &searchMemo{
		visited: map[relation]searchResults{},
		history: lo.Ternary(trace, []*searchCall{}, nil),
	}
}

func (m *searchMemo) MarkVisited(params *relation) searchStatus {
	results, ok := m.visited[*params]

	switch {
	case !ok:
		m.visited[*params] = nil
		m.trace(params, searchStatusPending)

		return searchStatusNew
	case results == nil:
		return searchStatusPending
	default:
		return searchStatusComplete
	}
}

func (m *searchMemo) MarkComplete(params *relation, results searchResults) {
	m.visited[*params] = results
	m.trace(params, searchStatusComplete)
}

func (m *searchMemo) Status(params *relation) searchStatus {
	results, ok := m.visited[*params]

	switch {
	case !ok:
		return searchStatusNew
	case results == nil:
		return searchStatusPending
	default:
		return searchStatusComplete
	}
}

func (m *searchMemo) Trace() []string {
	if m.history == nil {
		return []string{}
	}

	callstack := []string{}

	return lo.Map(m.history, func(c *searchCall, _ int) string {
		call := c.String()
		result := c.status.String()

		if len(callstack) > 0 && callstack[len(callstack)-1] == call && c.status != searchStatusPending {
			callstack = callstack[:len(callstack)-1]
		}

		s := fmt.Sprintf("%s%s = %s", strings.Repeat("  ", len(callstack)), call, result)

		if c.status == searchStatusPending {
			callstack = append(callstack, call)
		}

		return s
	})
}

func (m *searchMemo) trace(params *relation, status searchStatus) {
	if m.history != nil {
		m.history = append(m.history, &searchCall{params, status})
	}
}
