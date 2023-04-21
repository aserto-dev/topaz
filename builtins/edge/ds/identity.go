package ds

import (
	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	"github.com/aserto-dev/topaz/directory"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/rs/zerolog"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
	"github.com/pkg/errors"
)

// RegisterIdentity - ds.identity
//
// get user key for identity
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

			user, err := directory.GetIdentityV2(client, bctx.Context, a.Key)
			switch {
			case errors.Is(err, aerr.ErrDirectoryObjectNotFound):
				if !IsValidID(a.Key) {
					return nil, err
				}
			case err != nil:
				traceError(&bctx, fnName, err)
				return nil, err

			default:
				return ast.StringTerm(user.Key), nil
			}

			return nil, aerr.ErrDirectoryObjectNotFound
		}
}
