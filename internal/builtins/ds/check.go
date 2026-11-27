package ds

import (
	"github.com/aserto-dev/topaz/api/directory/v3/reader"
	"github.com/aserto-dev/topaz/internal/builtins"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/types"
	"google.golang.org/protobuf/proto"
)

// RegisterCheck - ds.check.
func RegisterCheck(fnName string, dr func() (reader.ReaderClient, error)) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.B),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args reader.CheckRequest

			dsr, err := dr()
			if err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			if err := ast.As(op1.Value, &args); err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			if proto.Equal(&args, &reader.CheckRequest{}) {
				return checkTemplate()
			}

			resp, err := dsr.Check(bctx.Context, &args)
			if err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			return ast.BooleanTerm(resp.GetCheck()), nil
		}
}

//nolint:goconst
func checkTemplate() (*ast.Term, error) {
	t := map[string]any{
		"object_type":  "",
		"object_id":    "",
		"relation":     "",
		"subject_type": "",
		"subject_id":   "",
		"trace":        false,
	}

	val, err := ast.InterfaceToValue(t)
	if err != nil {
		return nil, err
	}

	return ast.NewTerm(val), nil
}
