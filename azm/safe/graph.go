package safe

import (
	"github.com/aserto-dev/topaz/api/directory/pkg/derr"
	dsc3 "github.com/aserto-dev/topaz/api/directory/v4"
	dsr3 "github.com/aserto-dev/topaz/api/directory/v4/reader"
	"github.com/aserto-dev/topaz/azm/cache"
	"github.com/aserto-dev/topaz/azm/model"
)

type SafeGetGraph struct {
	*dsr3.GraphRequest
}

func GetGraph(i *dsr3.GraphRequest) *SafeGetGraph {
	return &SafeGetGraph{i}
}

func (i *SafeGetGraph) Object() *dsc3.ObjectIdentifier {
	return &dsc3.ObjectIdentifier{
		ObjectType: i.ObjectType,
		ObjectId:   i.ObjectId,
	}
}

func (i *SafeGetGraph) Subject() *dsc3.ObjectIdentifier {
	return &dsc3.ObjectIdentifier{
		ObjectType: i.SubjectType,
		ObjectId:   i.SubjectId,
	}
}

func (i *SafeGetGraph) RelationRef() *model.RelationRef {
	return &model.RelationRef{
		Object:   model.ObjectName(i.ObjectType),
		Relation: model.RelationName(i.Relation),
	}
}

func (i *SafeGetGraph) Validate(mc *cache.Cache) error {
	if i == nil || i.GraphRequest == nil {
		return derr.ErrInvalidRequest.Msg("get_graph")
	}

	// Object ID can be optional, hence the use of an ObjectSelector.
	if err := ObjectSelector(i.Object()).Validate(mc); err != nil {
		return err
	}

	// Subject ID can be option, hence the use of an ObjectSelector.
	if err := ObjectSelector(i.Subject()).Validate(mc); err != nil {
		return err
	}

	// Validate relation.
	if err := RelationIdentifier(i.RelationRef()).Validate(AsEither, mc); err != nil {
		return err
	}

	// Either object or subject must be specified in full.
	objErr := ObjectIdentifier(i.Object()).Validate(mc)
	subjErr := ObjectIdentifier(i.Subject()).Validate(mc)

	if objErr != nil && subjErr != nil {
		return derr.ErrInvalidArgument.Msg("object_id or subject_id must be specified")
	}

	return nil
}
