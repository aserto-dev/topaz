package dir

import (
	"github.com/aserto-dev/aserto-grpc/grpcutil"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/aserto-dev/topaz/directory"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// RegisterIsManagerOf - returns boolean indicating identity id1 is
// in transitive management chain of identity id2.
func RegisterIsManagerOf(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin2) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.S, types.S), types.B),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, a, b *ast.Term) (*ast.Term, error) {
			var (
				tenantID = grpcutil.ExtractTenantID(bctx.Context)
				mgr      string
				emp      string
			)

			if err := ast.As(a.Value, &mgr); err != nil {
				return nil, err
			}

			if err := ast.As(b.Value, &emp); err != nil {
				return nil, err
			}

			ds, err := dr.GetDirectory(bctx.Context, tenantID)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get directory")
			}

			isManagerOf := isManagerOf(logger, ds, tenantID, mgr, emp)

			return ast.BooleanTerm(isManagerOf), nil
		}
}

// isManagerOf - returns boolean indicating if identity id1 (role of manager)
// resides within the management chain of identity id2 (role of employee).
func isManagerOf(logger *zerolog.Logger, ds directory.Directory, tenantID, mgr, emp string) bool {
	logger.Trace().Str("tenantID", tenantID).Str("mgr", mgr).Str("emp", emp).Msg("IsManagerOf")

	var (
		empUser *api.User
		mgrUser *api.User
		err     error
	)

	empUser, err = ds.GetUserFromIdentity(tenantID, emp)
	if err != nil {
		logger.Error().Err(err).Interface("emp", emp).Msg("GetUserFromIdentity failed")
		return false
	}

	mgrUser, err = ds.GetUserFromIdentity(tenantID, mgr)
	if err != nil {
		logger.Error().Err(err).Interface("mgr", mgr).Msg("GetUserFromIdentity failed")
		return false
	}

	current := empUser

	for {
		nextAttr, ok := current.Attributes.Properties.Fields[manager]
		if !ok {
			logger.Error().Interface("current", current).Msg("no manager attribute found")
			break
		}

		mgrIdentity := nextAttr.GetStringValue()
		if mgrIdentity == "" {
			logger.Error().Interface("current", current).Msg("manager attribute is empty")
			break
		}

		nextUser, err := ds.GetUserFromIdentity(tenantID, mgrIdentity)
		if err != nil {
			logger.Error().Err(err).Interface("mgrIdentity", mgrIdentity).Msg("no manager attribute found")
			break
		}

		if nextUser.Id == mgrUser.Id {
			return true
		}

		current = nextUser
	}

	return false
}
