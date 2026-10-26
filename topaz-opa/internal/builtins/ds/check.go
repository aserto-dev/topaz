package ds

import (
	"errors"

	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/topaz-opa/internal/builtins"
	"github.com/aserto-dev/topaz/topaz-opa/internal/errs"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/types"
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

// registerCheck - ds.check.
func registerCheck(fnName string, dr func() (reader.ReaderClient, error)) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.B),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args reader.CheckRequest

			dsr, err := dr()
			if err != nil && errors.Is(err, errs.ErrTopazPluginDisabled) {
				return nil, err
			}

			if err := ast.As(op1.Value, &args); err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			if proto.Equal(&args, &reader.CheckRequest{}) {
				return ast.StringTerm(dsCheckHelp), nil
			}

			resp, err := dsr.Check(bctx.Context, &args)
			if err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			return ast.BooleanTerm(resp.GetCheck()), nil
		}
}
