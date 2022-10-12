package ds

import (
	"context"

	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	v2 "github.com/aserto-dev/go-directory/aserto/directory/common/v2"
	ds2 "github.com/aserto-dev/go-directory/aserto/directory/v2"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/rs/zerolog"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
	"github.com/pkg/errors"
)

// RegisterIdentity - ds.identity
//
// get user id for identity
//
// ds.identity({
//     "key": ""
// })
//
func RegisterIdentity(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: false,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {

			type args struct {
				Key string `json:"key"`
			}

			var a args
			if err := ast.As(op1.Value, &a); err != nil {
				return nil, err
			}

			if (args{}) == a {
				return help(fnName, args{})
			}

			client, err := dr.GetDS(bctx.Context)
			if err != nil {
				return nil, errors.Wrapf(err, "get directory client")
			}

			uid, err := getIdentityV2(bctx.Context, client, a.Key)
			switch {
			case errors.Is(err, aerr.ErrDirectoryObjectNotFound):
				if !IsValidID(a.Key) {
					return nil, err
				}
				uid = a.Key

			case err != nil:
				return nil, err

			default:
				return ast.StringTerm(uid), nil
			}

			// fallthrough check identity == uid,
			// validate existence of user object directly
			user, err := getUserV2(bctx.Context, client, uid)
			if err == nil {
				return ast.StringTerm(user.Result.Id), nil
			}

			return nil, aerr.ErrDirectoryObjectNotFound
		}
}

func getIdentityV2(ctx context.Context, client ds2.DirectoryClient, identity string) (string, error) {

	identityResp, err := client.GetObject(ctx, &ds2.GetObjectRequest{
		Param: &v2.ObjectIdentifier{
			Type: StrPrt("identity"),
			Key:  &identity,
		},
	})
	if err != nil {
		return "", err
	}

	if identityResp.Result == nil {
		return "", aerr.ErrDirectoryObjectNotFound
	}

	iid := identityResp.Result.Id

	relResp, err := client.GetRelation(ctx, &ds2.GetRelationRequest{
		Param: &v2.RelationIdentifier{
			Object:   &v2.ObjectIdentifier{Type: StrPrt("identity"), Id: &iid},
			Relation: &v2.RelationTypeIdentifier{Name: StrPrt("identifier"), ObjectType: StrPrt("identity")},
			Subject:  &v2.ObjectIdentifier{Type: StrPrt("user")},
		},
	})
	if err != nil {
		return "", err
	}

	if relResp.Result == nil {
		return "", aerr.ErrDirectoryObjectNotFound
	}

	uid := relResp.Result.Relation

	return uid, nil
}

func getUserV2(ctx context.Context, client ds2.DirectoryClient, uid string) (*ds2.GetObjectResponse, error) {
	userResp, err := client.GetObject(ctx, &ds2.GetObjectRequest{
		Param: &v2.ObjectIdentifier{
			Id: &uid,
		},
	})
	if err != nil {
		return nil, err
	}

	if userResp == nil {
		return nil, aerr.ErrDirectoryObjectNotFound
	}

	return userResp, nil
}
