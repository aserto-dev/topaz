//nolint:dupl
package ds

// import (
// 	dsc3 "github.com/aserto-dev/topaz/api/directory/v4"
// 	dsr3 "github.com/aserto-dev/topaz/api/directory/v4/reader"
// 	"github.com/aserto-dev/topaz/azm/cache"
// 	"github.com/aserto-dev/topaz/azm/model"
// )

// type checkRelation struct {
// 	*dsr3.CheckRelationRequest
// }

// func CheckRelation(i *dsr3.CheckRelationRequest) *checkRelation {
// 	return &checkRelation{i}
// }

// func (i *checkRelation) Object() *dsc3.ObjectIdentifier {
// 	return &dsc3.ObjectIdentifier{
// 		ObjectType: i.ObjectType,
// 		ObjectId:   i.ObjectId,
// 	}
// }

// func (i *checkRelation) Subject() *dsc3.ObjectIdentifier {
// 	return &dsc3.ObjectIdentifier{
// 		ObjectType: i.SubjectType,
// 		ObjectId:   i.SubjectId,
// 	}
// }

// func (i *checkRelation) Validate(mc *cache.Cache) error {
// 	if i == nil || i.CheckRelationRequest == nil {
// 		return ErrInvalidRequest.Msg("check_relation")
// 	}

// 	if err := ObjectIdentifier(i.Object()).Validate(mc); err != nil {
// 		return err
// 	}

// 	if err := ObjectIdentifier(i.Subject()).Validate(mc); err != nil {
// 		return err
// 	}

// 	if !mc.RelationExists(model.ObjectName(i.ObjectType), model.RelationName(i.Relation)) {
// 		return ErrRelationNotFound.Msgf("%s%s%s", i.ObjectType, RelationSeparator, i.Relation)
// 	}

// 	return nil
// }
