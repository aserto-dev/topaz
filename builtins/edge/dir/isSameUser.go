package dir

import (
	"github.com/aserto-dev/aserto-grpc/grpcutil"
	"github.com/aserto-dev/topaz/directory"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// RegisterIsSameUser - returns boolean indicating identity id1 and id2
// reference the same user instance.
func RegisterIsSameUser(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin2) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.S, types.S), types.B),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, a, b *ast.Term) (*ast.Term, error) {
			var (
				tenantID = grpcutil.ExtractTenantID(bctx.Context)
				id1      string
				id2      string
			)

			if err := ast.As(a.Value, &id1); err != nil {
				return nil, err
			}

			if err := ast.As(b.Value, &id2); err != nil {
				return nil, err
			}

			ds, err := dr.GetDirectory(bctx.Context, tenantID)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get directory")
			}

			isSameUser := isSameUser(logger, ds, tenantID, id1, id2)

			return ast.BooleanTerm(isSameUser), nil
		}
}

// isSameUser - return boolean indicating if identity id1 and id2,
// refer to the same user instance.
func isSameUser(logger *zerolog.Logger, ds directory.Directory, tenantID, id1, id2 string) bool {
	logger.Trace().Str("tenantID", tenantID).Str("id1", id1).Str("id2", id2).Msg("IsSameUser")

	ident1, err1 := ds.GetIdentity(tenantID, id1)
	ident2, err2 := ds.GetIdentity(tenantID, id2)

	return err1 == nil && err2 == nil && ident1 == ident2
}
