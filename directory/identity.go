package directory

import (
	"context"
	"errors"

	cerr "github.com/aserto-dev/errors"
	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	v2 "github.com/aserto-dev/go-directory/aserto/directory/common/v2"
	ds2 "github.com/aserto-dev/go-directory/aserto/directory/reader/v2"
	"github.com/aserto-dev/go-directory/pkg/derr"
)

func GetIdentityV2(client ds2.ReaderClient, ctx context.Context, identity string) (*v2.Object, error) {
	identityString := "identity"
	obj := v2.ObjectIdentifier{Type: &identityString, Key: &identity}

	relationString := "identifier"
	subjectType := "user"
	withObjects := true

	relResp, err := client.GetRelation(ctx, &ds2.GetRelationRequest{
		Param: &v2.RelationIdentifier{
			Object:   &obj,
			Relation: &v2.RelationTypeIdentifier{Name: &relationString, ObjectType: &identityString},
			Subject:  &v2.ObjectIdentifier{Type: &subjectType},
		},
		WithObjects: &withObjects,
	})
	switch {
	case err != nil && errors.Is(cerr.UnwrapAsertoError(err), derr.ErrNotFound):
		return nil, aerr.ErrDirectoryObjectNotFound
	case err != nil:
		return nil, err

	case relResp.Results == nil:
		return nil, aerr.ErrDirectoryObjectNotFound

	case len(relResp.Objects) == 0:
		return nil, aerr.ErrDirectoryObjectNotFound.Msg("no objects found in relation")
	}

	return relResp.Objects[*relResp.Results[0].Subject.Id], nil
}
