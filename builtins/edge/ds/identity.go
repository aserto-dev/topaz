package ds

import (
	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	"github.com/aserto-dev/topaz/directory"
	"github.com/aserto-dev/topaz/resolvers"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// RegisterIdentity - ds.identity - get user id (key) for identity
//
// v3 (latest) request format:
//
//	ds.identity({
//		"id": ""
//	})
//
// v2 request format:
//
//	ds.identity({
//		"key": ""
//	})
func RegisterIdentity(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {

			var args struct {
				ID  string `json:"id"`
				Key string `json:"key"`
			}

			if err := ast.As(op1.Value, &args); err != nil {
				return nil, err
			}

			if args.ID == "" && args.Key != "" {
				args.ID = args.Key
			}

			if args.ID == "" && args.Key == "" {
				type argsV3 struct {
					ID string `json:"id"`
				}
				return help(fnName, argsV3{})
			}

			client, err := dr.GetDS(bctx.Context)
			if err != nil {
				return nil, errors.Wrapf(err, "get directory client")
			}

			user, err := directory.GetIdentityV2(bctx.Context, client, args.ID)
			switch {
			case errors.Is(err, aerr.ErrDirectoryObjectNotFound):
				return nil, err
			case err != nil:
				traceError(&bctx, fnName, err)
				return nil, err
			default:
				return ast.StringTerm(user.Id), nil
			}
		}
}
