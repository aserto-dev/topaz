package dir

import (
	"bytes"
	"fmt"

	"github.com/aserto-dev/aserto-grpc/grpcutil"
	"github.com/aserto-dev/go-eds/pkg/pb"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/aserto-dev/topaz/directory"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
)

const manager string = "manager"

// RegisterManagerOf - return managers user object of identity.
func RegisterManagerOf(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.S), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, a *ast.Term) (*ast.Term, error) {
			var (
				tenantID = grpcutil.ExtractTenantID(bctx.Context)
				identity string
			)

			if err := ast.As(a.Value, &identity); err != nil {
				return nil, err
			}

			ds, err := dr.GetDirectory(bctx.Context, tenantID)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get directory")
			}

			user, ok := managerOf(logger, ds, tenantID, identity)
			if !ok {
				return nil, fmt.Errorf("cannot determine manager of %s", identity)
			}

			buf := new(bytes.Buffer)
			if err := pb.ProtoToBuf(buf, user); err != nil {
				return nil, err
			}

			v, err := ast.ValueFromReader(buf)
			if err != nil {
				return nil, err
			}

			return ast.NewTerm(v), nil
		}
}

// managerOf returns the user object of the manager of the given identity.
func managerOf(logger *zerolog.Logger, ds directory.Directory, tenantID, ident string) (*api.User, bool) {
	logger.Trace().Str("tenantID", tenantID).Str("ident", ident).Msg("ManagerOf")

	empUser, err := ds.GetUserFromIdentity(tenantID, ident)
	if err != nil {
		return nil, false
	}

	mgrAttr, ok := empUser.Attributes.Properties.Fields[manager]
	if !ok {
		return &api.User{}, ok
	}

	mgrUID := mgrAttr.GetStringValue()

	mgrUser, err := ds.GetUserFromIdentity(tenantID, mgrUID)
	if err != nil {
		return &api.User{}, false
	}

	return mgrUser, true
}
