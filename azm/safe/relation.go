package safe

import (
	"hash"
	"hash/fnv"
	"strconv"

	"github.com/aserto-dev/topaz/api/directory/pkg/derr"
	dsc3 "github.com/aserto-dev/topaz/api/directory/v4"
	dsr3 "github.com/aserto-dev/topaz/api/directory/v4/reader"
	"github.com/aserto-dev/topaz/azm/cache"
	"github.com/aserto-dev/topaz/azm/model"
)

// SafeRelation identifier.
type SafeRelation struct {
	*dsc3.RelationIdentifier

	HasSubjectRelation bool
}

func Relation(i *dsc3.RelationIdentifier) *SafeRelation { return &SafeRelation{i, true} }

type SafeRelationIdentifier struct {
	*model.RelationRef
}

func RelationIdentifier(i *model.RelationRef) *SafeRelationIdentifier {
	return &SafeRelationIdentifier{i}
}

// SafeRelations selector.
type SafeRelations struct {
	*SafeRelation
}

func GetRelation(i *dsr3.GetRelationRequest) *SafeRelations {
	return &SafeRelations{
		&SafeRelation{
			RelationIdentifier: &dsc3.RelationIdentifier{
				ObjectType:      i.GetObjectType(),
				ObjectId:        i.GetObjectId(),
				Relation:        i.GetRelation(),
				SubjectType:     i.GetSubjectType(),
				SubjectId:       i.GetSubjectId(),
				SubjectRelation: i.GetSubjectRelation(),
			},
			HasSubjectRelation: true,
		},
	}
}

func GetRelations(i *dsr3.ListRelationsRequest) *SafeRelations {
	return &SafeRelations{
		&SafeRelation{
			RelationIdentifier: &dsc3.RelationIdentifier{
				ObjectType:      i.GetObjectType(),
				ObjectId:        i.GetObjectId(),
				Relation:        i.GetRelation(),
				SubjectType:     i.GetSubjectType(),
				SubjectId:       i.GetSubjectId(),
				SubjectRelation: i.GetSubjectRelation(),
			},
			HasSubjectRelation: i.GetSubjectRelation() != "" || i.GetWithEmptySubjectRelation(),
		},
	}
}

func (i *SafeRelation) Object() *dsc3.ObjectIdentifier {
	return &dsc3.ObjectIdentifier{
		ObjectType: i.GetObjectType(),
		ObjectId:   i.GetObjectId(),
	}
}

func (i *SafeRelation) Subject() *dsc3.ObjectIdentifier {
	return &dsc3.ObjectIdentifier{
		ObjectType: i.GetSubjectType(),
		ObjectId:   i.GetSubjectId(),
	}
}

func (i *SafeRelation) Validate(mc *cache.Cache) error {
	if i == nil || i.RelationIdentifier == nil {
		return derr.ErrInvalidRelation.Msg("relation not set (nil)")
	}

	if IsNotSet(i.GetRelation()) {
		return derr.ErrInvalidRelation.Msg("relation")
	}

	if err := ObjectIdentifier(i.Object()).Validate(mc); err != nil {
		return err
	}

	if err := ObjectIdentifier(i.Subject()).Validate(mc); err != nil {
		return err
	}

	if mc == nil {
		return nil
	}

	if !mc.RelationExists(model.ObjectName(i.GetObjectType()), model.RelationName(i.GetRelation())) {
		return derr.ErrRelationTypeNotFound.Msg(i.GetObjectType() + ":" + i.GetRelation())
	}

	if IsSet(i.GetSubjectRelation()) {
		if !mc.RelationExists(model.ObjectName(i.GetSubjectType()), model.RelationName(i.GetSubjectRelation())) {
			return derr.ErrRelationTypeNotFound.Msg(i.GetSubjectType() + ":" + i.GetSubjectRelation())
		}
	}

	return mc.ValidateRelation(i.RelationIdentifier)
}

func (i *SafeRelations) Validate(mc *cache.Cache) error {
	if i == nil || i.RelationIdentifier == nil {
		return derr.ErrInvalidRelation.Msg("relation not set (nil)")
	}

	if err := ObjectSelector(i.Object()).Validate(mc); err != nil {
		return err
	}

	if err := ObjectSelector(i.Subject()).Validate(mc); err != nil {
		return err
	}

	if err := i.validateRelation(mc); err != nil {
		return err
	}

	return i.validateSubjectRelation(mc)
}

func (i *SafeRelation) Hash() string {
	h := fnv.New64a()
	h.Reset()

	if err := i.writeHash(h); err != nil {
		return DefaultHash
	}

	return strconv.FormatUint(h.Sum64(), 10)
}

func (i *SafeRelation) relationExists(objType, relation string, mc *cache.Cache) error {
	if mc != nil && !mc.RelationExists(model.ObjectName(objType), model.RelationName(relation)) {
		return derr.ErrRelationTypeNotFound.Msg(objType + ":" + relation)
	}

	return nil
}

func (i *SafeRelation) writeHash(h hash.Hash64) error {
	if i == nil || i.RelationIdentifier == nil {
		return nil
	}

	if _, err := h.Write([]byte(i.GetObjectId())); err != nil {
		return err
	}

	if _, err := h.Write([]byte(i.GetObjectType())); err != nil {
		return err
	}

	if _, err := h.Write([]byte(i.GetRelation())); err != nil {
		return err
	}

	if _, err := h.Write([]byte(i.GetSubjectId())); err != nil {
		return err
	}

	if _, err := h.Write([]byte(i.GetSubjectType())); err != nil {
		return err
	}

	if _, err := h.Write([]byte(i.GetSubjectRelation())); err != nil {
		return err
	}

	return nil
}

func (i *SafeRelation) validateRelation(mc *cache.Cache) error {
	if !IsSet(i.GetRelation()) {
		return nil
	}

	if IsNotSet(i.GetObjectType()) {
		return derr.ErrInvalidRelation.Msg("object type not set")
	}

	return i.relationExists(i.GetObjectType(), i.GetRelation(), mc)
}

func (i *SafeRelation) validateSubjectRelation(mc *cache.Cache) error {
	if !IsSet(i.GetSubjectRelation()) {
		return nil
	}

	if IsNotSet(i.GetSubjectType()) {
		return derr.ErrInvalidRelation.Msg("subject type not set")
	}

	return i.relationExists(i.GetSubjectType(), i.GetSubjectRelation(), mc)
}

type RelationScope int

const (
	AsRelation RelationScope = iota
	AsPermission
	AsEither
)

func (r *SafeRelationIdentifier) Validate(scope RelationScope, mc *cache.Cache) error {
	if r == nil || r.RelationRef == nil {
		return derr.ErrInvalidRelation.Msg("relation not set (nil)")
	}

	if r.Object == "" {
		return derr.ErrInvalidRelation.Msg("object")
	}

	if r.Relation == "" {
		return derr.ErrInvalidRelation.Msg("relation")
	}

	if mc == nil {
		return nil
	}

	return r.validateExistence(scope, mc)
}

func (r *SafeRelationIdentifier) validateExistence(scope RelationScope, mc *cache.Cache) error {
	switch scope {
	case AsRelation:
		if !mc.RelationExists(r.Object, r.Relation) {
			return derr.ErrRelationTypeNotFound.Msgf("relation: %s", r)
		}
	case AsPermission:
		if !mc.PermissionExists(r.Object, r.Relation) {
			return derr.ErrPermissionNotFound.Msgf("permission: %s", r)
		}
	case AsEither:
		if !mc.RelationExists(r.Object, r.Relation) && !mc.PermissionExists(r.Object, r.Relation) {
			return derr.ErrRelationTypeNotFound.Msgf("relation: %s", r)
		}
	}

	return nil
}
