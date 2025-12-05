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

const dsCheckHelp string = `ds.check({
	"object_type": "",
	"object_id": "",
	"relation": "",
	"subject_type": "",
	"subject_id": "",
	"trace": false
})`

// RegisterCheck - ds.check.
func RegisterCheck(logger *zerolog.Logger, fnName string, dr reader.ReaderClient) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.B),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args reader.CheckRequest

			if err := ast.As(op1.Value, &args); err != nil {
				return nil, err
			}

			if proto.Equal(&args, &reader.CheckRequest{}) {
				return ast.StringTerm(dsCheckHelp), nil
			}

			resp, err := dr.Check(bctx.Context, &args)
			if err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			return ast.BooleanTerm(resp.GetCheck()), nil
		}
}
