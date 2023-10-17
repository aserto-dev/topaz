package directory

import (
	"context"
	"errors"
	"strings"

	cerr "github.com/aserto-dev/errors"
	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	dsc2 "github.com/aserto-dev/go-directory/aserto/directory/common/v2"
	dsr2 "github.com/aserto-dev/go-directory/aserto/directory/reader/v2"
	"github.com/aserto-dev/go-directory/pkg/derr"
)

func GetIdentityV2(client dsr2.ReaderClient, ctx context.Context, identity string) (*dsc2.Object, error) {
	identityString := "identity"
	obj := dsc2.ObjectIdentifier{Type: &identityString, Key: &identity}

	relationString := "identifier"
	subjectType := "user"
	withObjects := true

	relResp, err := client.GetRelation(ctx, &dsr2.GetRelationRequest{
		Param: &dsc2.RelationIdentifier{
			Object:   &obj,
			Relation: &dsc2.RelationTypeIdentifier{Name: &relationString, ObjectType: &identityString},
			Subject:  &dsc2.ObjectIdentifier{Type: &subjectType},
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

	for k, v := range relResp.Objects {
		if strings.HasPrefix(k, "user:") {
			return v, nil
		}
	}

	return nil, aerr.ErrDirectoryObjectNotFound
}
