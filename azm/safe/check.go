package safe

import (
	"github.com/aserto-dev/topaz/azm/cache"
	"github.com/aserto-dev/topaz/azm/model"

	"github.com/aserto-dev/topaz/api/directory/pkg/derr"
	dsc3 "github.com/aserto-dev/topaz/api/directory/v4"
	dsr3 "github.com/aserto-dev/topaz/api/directory/v4/reader"
)

type SafeCheck struct {
	*dsr3.CheckRequest
}

func Check(i *dsr3.CheckRequest) *SafeCheck {
	return &SafeCheck{i}
}

func (i *SafeCheck) Object() *dsc3.ObjectIdentifier {
	return &dsc3.ObjectIdentifier{
		ObjectType: i.ObjectType,
		ObjectId:   i.ObjectId,
	}
}

func (i *SafeCheck) Subject() *dsc3.ObjectIdentifier {
	return &dsc3.ObjectIdentifier{
		ObjectType: i.SubjectType,
		ObjectId:   i.SubjectId,
	}
}

func (i *SafeCheck) Validate(mc *cache.Cache) error {
	if i == nil || i.CheckRequest == nil {
		return derr.ErrInvalidRequest.Msg("check")
	}

	if err := ObjectIdentifier(i.Object()).Validate(mc); err != nil {
		return err.Msg("object_type")
	}

	if err := ObjectIdentifier(i.Subject()).Validate(mc); err != nil {
		return err.Msg("subject_type")
	}

	rr := &model.RelationRef{
		Object:   model.ObjectName(i.ObjectType),
		Relation: model.RelationName(i.Relation),
	}

	return RelationIdentifier(rr).Validate(AsEither, mc)
}
