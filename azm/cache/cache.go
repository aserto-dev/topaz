package cache

import (
	"sync/atomic"

	"github.com/aserto-dev/topaz/api/directory/pkg/derr"
	dsc "github.com/aserto-dev/topaz/api/directory/v4"
	"github.com/aserto-dev/topaz/azm/mempool"
	"github.com/aserto-dev/topaz/azm/model"
	"github.com/aserto-dev/topaz/azm/model/diff"
	stts "github.com/aserto-dev/topaz/azm/stats"
	"github.com/samber/lo"
)

type (
	ObjectName   = model.ObjectName
	RelationName = model.RelationName
)

type Cache struct {
	model    atomic.Pointer[model.Model]
	relsPool *mempool.RelationsPool
}

// New, create new model cache instance.
func New(m *model.Model) *Cache {
	cache := &Cache{
		relsPool: mempool.NewRelationsPool(),
	}

	cache.model.Store(m)

	return cache
}

// UpdateModel, swaps the cache model instance.
func (c *Cache) UpdateModel(m *model.Model) error {
	c.model.Store(m)
	return nil
}

func (c *Cache) CanUpdate(other *model.Model, stats *stts.Stats) error {
	return diff.CanUpdateModel(c.model.Load(), other, stats)
}

// ObjectExists, checks if given object type name exists in the model cache.
func (c *Cache) ObjectExists(on ObjectName) bool {
	_, ok := c.model.Load().Objects[on]
	return ok
}

// RelationExists, checks if given relation type, for the given object type, exists in the model cache.
func (c *Cache) RelationExists(on ObjectName, rn RelationName) bool {
	if obj, ok := c.model.Load().Objects[on]; ok {
		_, ok := obj.Relations[rn]
		return ok
	}

	return false
}

// PermissionExists, checks if given permission, for the given object type, exists in the model cache.
func (c *Cache) PermissionExists(on ObjectName, pn RelationName) bool {
	if obj, ok := c.model.Load().Objects[on]; ok {
		_, ok := obj.Permissions[pn]
		return ok
	}

	return false
}

func (c *Cache) Metadata() *model.Metadata {
	return c.model.Load().Metadata
}

func (c *Cache) ValidateRelation(relation *dsc.RelationIdentifier) error {
	return c.model.Load().ValidateRelation(
		ObjectName(relation.GetObjectType()),
		model.ObjectID(relation.GetObjectId()),
		RelationName(relation.GetRelation()),
		ObjectName(relation.GetSubjectType()),
		model.ObjectID(relation.GetSubjectId()),
		RelationName(relation.GetSubjectRelation()),
	)
}

// AssignableRelations returns the set of relations that can occur between a given object type
// and a subject type, optionally with a subject relation.
//
// If more than one subject relation is provided, AssignableRelations returns relations that match any
// of the given relations. For example, if the manifest has:
//
// types:
//
//	tenant:
//	  relations:
//	    admin: group#member
//	    guest: group#guest
//
// Then AssignableRelations("tenant", "group", "member", "guest") returns ["admin", "guest"].
func (c *Cache) AssignableRelations(on, sn ObjectName, sr ...RelationName) ([]RelationName, error) {
	if err := c.validateTypes(on, sn, sr...); err != nil {
		return nil, err
	}

	matches := lo.PickBy(c.model.Load().Objects[on].Relations, func(rn RelationName, r *model.Relation) bool {
		for _, ref := range r.Union {
			if ref.Object != sn {
				// type mismatch
				continue
			}

			switch {
			case ref.IsDirect(), ref.IsWildcard():
				if len(sr) == 0 {
					return true
				}
			case ref.IsSubject() && lo.Contains(sr, ref.Relation):
				return true
			}
		}

		return false
	})

	return lo.Keys(matches), nil
}

// AvailablePermissions returns the set of permissions that a given subject type can have on an object type,
// optionally with a subject relation.
//
// If more than one subject relation is provided, AvailablePermissions returns permissions that match any
// of the given relations.
func (c *Cache) AvailablePermissions(on, sn ObjectName, sr ...RelationName) ([]RelationName, error) {
	if err := c.validateTypes(on, sn, sr...); err != nil {
		return nil, err
	}

	matches := lo.PickBy(c.model.Load().Objects[on].Permissions, func(pn RelationName, p *model.Permission) bool {
		if len(sr) == 0 {
			return lo.Contains(p.SubjectTypes, sn)
		}

		for _, srn := range sr {
			if lo.Contains(p.Intermediates, model.RelationRef{Object: sn, Relation: srn}) {
				return true
			}
		}

		return false
	})

	return lo.Keys(matches), nil
}

func (c *Cache) validateTypes(on, sn ObjectName, sr ...RelationName) error {
	if !c.ObjectExists(on) {
		return derr.ErrObjectTypeNotFound.Msg(on.String())
	}

	if !c.ObjectExists(sn) {
		return derr.ErrObjectTypeNotFound.Msg(sn.String())
	}

	for _, srel := range sr {
		if !c.RelationExists(sn, srel) {
			return derr.ErrRelationTypeNotFound.Msgf("%s#%s", sn, sr)
		}
	}

	return nil
}
