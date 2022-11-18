package ds

import (
	"context"

	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	dsc "github.com/aserto-dev/go-directory/aserto/directory/common/v2"
	dsr "github.com/aserto-dev/go-directory/aserto/directory/reader/v2"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/proto"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
	"github.com/pkg/errors"
)

// RegisterIdentity - ds.identity
//
// get user id for identity
//
//	ds.identity({
//		"key": ""
//	})
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

func getIdentityV2(ctx context.Context, client dsr.ReaderClient, identity string) (string, error) {

	identityResp, err := client.GetObject(ctx, &dsr.GetObjectRequest{
		Param: &dsc.ObjectIdentifier{
			Type: proto.String("identity"),
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

	relResp, err := client.GetRelation(ctx, &dsr.GetRelationRequest{
		Param: &dsc.RelationIdentifier{
			Object:   &dsc.ObjectIdentifier{Type: proto.String("identity"), Id: &iid},
			Relation: &dsc.RelationTypeIdentifier{Name: proto.String("identifier"), ObjectType: proto.String("identity")},
			Subject:  &dsc.ObjectIdentifier{Type: proto.String("user")},
		},
	})
	if err != nil {
		return "", err
	}

	if relResp.Results == nil || len(relResp.Results) == 0 {
		return "", aerr.ErrDirectoryObjectNotFound
	}

	uid := *relResp.Results[0].Subject.Id

	return uid, nil
}

func getUserV2(ctx context.Context, client dsr.ReaderClient, uid string) (*dsr.GetObjectResponse, error) {
	userResp, err := client.GetObject(ctx, &dsr.GetObjectRequest{
		Param: &dsc.ObjectIdentifier{
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
