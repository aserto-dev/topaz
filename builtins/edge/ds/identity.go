package ds

import (
	"github.com/aserto-dev/topaz/directory"
	"github.com/aserto-dev/topaz/resolvers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"

	"github.com/rs/zerolog"
)

// RegisterIdentity - ds.identity - get user id (key) for identity
//
//	ds.identity({
//		"id": ""
//	})
func RegisterIdentity(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args struct {
				ID string `json:"id"`
			}

			if err := ast.As(op1.Value, &args); err != nil {
				traceError(&bctx, fnName, err)
				return nil, err
			}

			if args.ID == "" {
				type argsV3 struct {
					ID string `json:"id"`
				}
				return help(fnName, argsV3{})
			}

			user, err := directory.GetIdentityV2(bctx.Context, dr.GetDS(), args.ID)
			switch {
			case status.Code(err) == codes.NotFound:
				traceError(&bctx, fnName, err)
				astVal, err := ast.InterfaceToValue(map[string]any{})
				if err != nil {
					return nil, err
				}
				return ast.NewTerm(astVal), nil
			case err != nil:
				return nil, err
			default:
				return ast.StringTerm(user.Id), nil
			}
		}
}
