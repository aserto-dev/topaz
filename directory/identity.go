package directory

import (
	"context"
	"errors"
	"strings"

	cerr "github.com/aserto-dev/errors"
	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	dsc3 "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/go-directory/pkg/derr"
)

func GetIdentityV2(ctx context.Context, client dsr3.ReaderClient, identity string) (*dsc3.Object, error) {
	relResp, err := client.GetRelation(ctx, &dsr3.GetRelationRequest{
		ObjectType:  "identity",
		ObjectId:    identity,
		Relation:    "identifier",
		SubjectType: "user",
		WithObjects: true,
	})

	switch {
	case err != nil && errors.Is(cerr.UnwrapAsertoError(err), derr.ErrNotFound):
		return nil, aerr.ErrDirectoryObjectNotFound
	case err != nil:
		return nil, err

	case relResp.Result == nil:
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
