package ds

import (
	"github.com/aserto-dev/topaz/api/directory/v3/reader"
	"github.com/aserto-dev/topaz/internal/builtins"
	"github.com/aserto-dev/topaz/topazd/directory"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/types"
)

// RegisterIdentity - ds.identity - get user id (key) for identity.
func RegisterIdentity(fnName string, dr func() (reader.ReaderClient, error)) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args struct {
				ID string `json:"id"`
			}

			dsr, err := dr()
			if err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			if err := ast.As(op1.Value, &args); err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			if args.ID == "" {
				return identityTemplate()
			}

			user, err := directory.ResolveIdentity(bctx.Context, dsr, args.ID)

			switch {
			case status.Code(err) == codes.NotFound:
				builtins.TraceError(&bctx, fnName, err)

				astVal, err := ast.InterfaceToValue(map[string]any{})
				if err != nil {
					return nil, err
				}

				return ast.NewTerm(astVal), nil
			case err != nil:
				return nil, err
			default:
				return ast.StringTerm(user.GetId()), nil
			}
		}
}

func identityTemplate() (*ast.Term, error) {
	t := map[string]any{
		"id": "",
	}

	val, err := ast.InterfaceToValue(t)
	if err != nil {
		return nil, err
	}

	return ast.NewTerm(val), nil
}
