package dir

import (
	"bytes"

	"github.com/aserto-dev/go-eds/pkg/pb"
	"github.com/aserto-dev/topaz/pkg/app/instance"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
)

// RegisterUser - convert user id into user object.
func RegisterUser(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.S), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, a *ast.Term) (*ast.Term, error) {
			var (
				instanceID = instance.ExtractID(bctx.Context)
				uid        string
			)

			if err := ast.As(a.Value, &uid); err != nil {
				return nil, err
			}

			ds, err := dr.GetDirectory(bctx.Context, instanceID)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get directory")
			}

			user, err := ds.GetUserFromIdentity(instanceID, uid)
			if err != nil {
				return nil, errors.Wrapf(err, "user not found [%s]", uid)
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
