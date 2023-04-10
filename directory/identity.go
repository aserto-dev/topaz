package directory

import (
	"context"
	"errors"
	"fmt"

	cerr "github.com/aserto-dev/errors"
	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	v2 "github.com/aserto-dev/go-directory/aserto/directory/common/v2"
	ds2 "github.com/aserto-dev/go-directory/aserto/directory/reader/v2"
	"github.com/aserto-dev/go-directory/pkg/derr"
	"google.golang.org/protobuf/proto"
)

func GetIdentityV2(client ds2.ReaderClient, ctx context.Context, identity string) (*v2.Object, error) {
	identityString := "identity"
	subjectType := "user"

	obj := v2.ObjectIdentifier{Type: &identityString, Key: &identity}

	relResp, err := client.GetRelation(ctx, &ds2.GetRelationRequest{
		Param: &v2.RelationIdentifier{
			Object:   &obj,
			Relation: &v2.RelationTypeIdentifier{Name: proto.String("identifier"), ObjectType: &identityString},
			Subject:  &v2.ObjectIdentifier{Type: &subjectType},
		},
		WithObjects: proto.Bool(true),
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

	return relResp.Objects[fmt.Sprintf("%s:%s", subjectType, *relResp.Results[0].Subject.Key)], nil
}
