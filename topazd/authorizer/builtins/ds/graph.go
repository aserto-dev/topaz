package ds

import (
	"bytes"

	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/topazd/authorizer/builtins"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/types"

	"github.com/rs/zerolog"
	"google.golang.org/protobuf/proto"
)

const dsGraphHelp string = `ds.graph({
	"object_type": "",
	"object_id": "",
	"relation": "",
	"subject_type": "",
	"subject_id": "",
	"subject_relation": "",
	"explain": false,
	"trace": false
})`

// RegisterGraph - ds.graph.
func RegisterGraph(logger *zerolog.Logger, fnName string, dr reader.ReaderClient) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args reader.GetGraphRequest

			if err := ast.As(op1.Value, &args); err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			if proto.Equal(&args, &reader.GetGraphRequest{}) {
				return ast.StringTerm(dsGraphHelp), nil
			}

			resp, err := dr.GetGraph(bctx.Context, &args)
			if err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			buf := new(bytes.Buffer)
			if len(resp.GetResults()) > 0 {
				if err := builtins.ProtoToBuf(buf, resp); err != nil {
					return nil, err
				}
			}

			v, err := ast.ValueFromReader(buf)
			if err != nil {
				return nil, err
			}

			return ast.NewTerm(v), nil
		}
}
