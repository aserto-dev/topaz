//nolint:dupl
package ds

import (
	"github.com/aserto-dev/azm/cache"
	"github.com/aserto-dev/azm/model"
	dsc "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsr "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
)

type checkRelation struct {
	*dsr.CheckRelationRequest
}

func CheckRelation(i *dsr.CheckRelationRequest) *checkRelation {
	return &checkRelation{i}
}

func (i *checkRelation) Object() *dsc.ObjectIdentifier {
	return &dsc.ObjectIdentifier{
		ObjectType: i.ObjectType,
		ObjectId:   i.ObjectId,
	}
}

func (i *checkRelation) Subject() *dsc.ObjectIdentifier {
	return &dsc.ObjectIdentifier{
		ObjectType: i.SubjectType,
		ObjectId:   i.SubjectId,
	}
}

func (i *checkRelation) Validate(mc *cache.Cache) error {
	if i == nil || i.CheckRelationRequest == nil {
		return ErrInvalidRequest.Msg("check_relation")
	}

	if err := ObjectIdentifier(i.Object()).Validate(mc); err != nil {
		return err
	}

	if err := ObjectIdentifier(i.Subject()).Validate(mc); err != nil {
		return err
	}

	if !mc.RelationExists(model.ObjectName(i.ObjectType), model.RelationName(i.Relation)) {
		return ErrRelationNotFound.Msgf("%s%s%s", i.ObjectType, RelationSeparator, i.Relation)
	}

	return nil
}
