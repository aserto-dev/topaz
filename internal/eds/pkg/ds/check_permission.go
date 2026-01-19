//nolint:dupl
package ds

import (
	"github.com/aserto-dev/azm/cache"
	"github.com/aserto-dev/azm/model"
	dsc3 "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
)

type checkPermission struct {
	*dsr3.CheckPermissionRequest
}

func CheckPermission(i *dsr3.CheckPermissionRequest) *checkPermission {
	return &checkPermission{i}
}

func (i *checkPermission) Object() *dsc3.ObjectIdentifier {
	return &dsc3.ObjectIdentifier{
		ObjectType: i.ObjectType,
		ObjectId:   i.ObjectId,
	}
}

func (i *checkPermission) Subject() *dsc3.ObjectIdentifier {
	return &dsc3.ObjectIdentifier{
		ObjectType: i.SubjectType,
		ObjectId:   i.SubjectId,
	}
}

func (i *checkPermission) Validate(mc *cache.Cache) error {
	if i == nil || i.CheckPermissionRequest == nil {
		return ErrInvalidRequest.Msg("check_permission")
	}

	if err := ObjectIdentifier(i.Object()).Validate(mc); err != nil {
		return err
	}

	if err := ObjectIdentifier(i.Subject()).Validate(mc); err != nil {
		return err
	}

	if !mc.PermissionExists(model.ObjectName(i.ObjectType), model.RelationName(i.Permission)) {
		return ErrPermissionNotFound.Msgf("%s%s%s", i.ObjectType, RelationSeparator, i.Permission)
	}

	return nil
}
