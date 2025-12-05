package ds

import (
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/builtins"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/types"

	"github.com/rs/zerolog"
	"google.golang.org/protobuf/proto"
)

const dsCheckPermissionHelp string = `ds.check_permission({
	"object_id": "",
	"object_type": "",
	"permission": "",
	"subject_id": "",
	"subject_type": "",
	"trace": false
})`

// RegisterCheckPermission - ds.check_permission (OBSOLETE).
func RegisterCheckPermission(logger *zerolog.Logger, fnName string, dr reader.ReaderClient) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.B),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args reader.CheckPermissionRequest

			if err := ast.As(op1.Value, &args); err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			if proto.Equal(&args, &reader.CheckPermissionRequest{}) {
				return ast.StringTerm(dsCheckPermissionHelp), nil
			}

			//nolint: staticcheck // SA1019: client.CheckPermission is deprecated
			resp, err := dr.CheckPermission(bctx.Context, &args)
			if err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			return ast.BooleanTerm(resp.GetCheck()), nil
		}
}
