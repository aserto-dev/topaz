//nolint:dupl
package ds

// import (
// 	dsc3 "github.com/aserto-dev/topaz/api/directory/v4"
// 	dsr3 "github.com/aserto-dev/topaz/api/directory/v4/reader"
// 	"github.com/aserto-dev/topaz/azm/cache"
// 	"github.com/aserto-dev/topaz/azm/model"
// )

// type checkPermission struct {
// 	*dsr3.CheckPermissionRequest
// }

// func CheckPermission(i *dsr3.CheckPermissionRequest) *checkPermission {
// 	return &checkPermission{i}
// }

// func (i *checkPermission) Object() *dsc3.ObjectIdentifier {
// 	return &dsc3.ObjectIdentifier{
// 		ObjectType: i.ObjectType,
// 		ObjectId:   i.ObjectId,
// 	}
// }

// func (i *checkPermission) Subject() *dsc3.ObjectIdentifier {
// 	return &dsc3.ObjectIdentifier{
// 		ObjectType: i.SubjectType,
// 		ObjectId:   i.SubjectId,
// 	}
// }

// func (i *checkPermission) Validate(mc *cache.Cache) error {
// 	if i == nil || i.CheckPermissionRequest == nil {
// 		return ErrInvalidRequest.Msg("check_permission")
// 	}

// 	if err := ObjectIdentifier(i.Object()).Validate(mc); err != nil {
// 		return err
// 	}

// 	if err := ObjectIdentifier(i.Subject()).Validate(mc); err != nil {
// 		return err
// 	}

// 	if !mc.PermissionExists(model.ObjectName(i.ObjectType), model.RelationName(i.Permission)) {
// 		return ErrPermissionNotFound.Msgf("%s%s%s", i.ObjectType, RelationSeparator, i.Permission)
// 	}

// 	return nil
// }
