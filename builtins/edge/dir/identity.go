package dir

import (
	"github.com/aserto-dev/topaz/pkg/app/instance"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/rs/zerolog"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
	"github.com/pkg/errors"
)

// RegisterIdentity - convert identity into user id.
func RegisterIdentity(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.S), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, a *ast.Term) (*ast.Term, error) {
			var (
				instanceID = instance.ExtractID(bctx.Context)
				ident      string
			)

			if err := ast.As(a.Value, &ident); err != nil {
				return nil, err
			}

			ds, err := dr.GetDirectory(bctx.Context, instanceID)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get directory")
			}

			uid, err := ds.GetIdentity(instanceID, ident)
			if err != nil {
				return nil, errors.Wrapf(err, "identity not found %s", ident)
			}

			return ast.StringTerm(uid), nil
		}
}
