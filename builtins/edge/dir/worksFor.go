package dir

import (
	"github.com/aserto-dev/aserto-grpc/grpcutil"
	"github.com/aserto-dev/topaz/directory"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
)

// RegisterWorksFor - returns boolean indicating identity A is in the
// transitive management chain of identity B. This this the inverse of IsManagerOf().
func RegisterWorksFor(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin2) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.S, types.S), types.B),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, a, b *ast.Term) (*ast.Term, error) {
			var (
				tenantID = grpcutil.ExtractTenantID(bctx.Context)
				emp      string
				mgr      string
			)

			if err := ast.As(a.Value, &emp); err != nil {
				return nil, err
			}

			if err := ast.As(b.Value, &mgr); err != nil {
				return nil, err
			}

			ds, err := dr.GetDirectory(bctx.Context, tenantID)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get directory")
			}

			worksFor := worksFor(logger, ds, tenantID, emp, mgr)

			return ast.BooleanTerm(worksFor), nil
		}
}

// worksFor - returns boolean indicating if identity id1 (role of employee)
// is working for identity id2 (role of manager).
func worksFor(logger *zerolog.Logger, ds directory.Directory, tenantID, id1, id2 string) bool {
	logger.Trace().Str("tenantID", tenantID).Str("id1", id1).Str("id2", id2).Msg("WorksFor")

	return isManagerOf(logger, ds, tenantID, id2, id1)
}
