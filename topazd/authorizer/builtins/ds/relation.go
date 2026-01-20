package ds

import (
	"bytes"

	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/topazd/authorizer/builtins"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/types"

	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const dsRelationHelp string = `ds.relation({
	"object_id": "",
	"object_type": "",
	"relation": "",
	"subject_id": "",
	"subject_relation": "",
	"subject_type": "",
	"with_objects": false
	})`

// RegisterRelation - ds.relation.
func RegisterRelation(logger *zerolog.Logger, fnName string, dr reader.ReaderClient) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args reader.GetRelationRequest

			if err := ast.As(op1.Value, &args); err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			if proto.Equal(&args, &reader.GetRelationRequest{}) {
				return ast.StringTerm(dsRelationHelp), nil
			}

			resp, err := dr.GetRelation(bctx.Context, &args)

			switch {
			case status.Code(err) == codes.NotFound:
				builtins.TraceError(&bctx, fnName, err)

				astVal, err := ast.InterfaceToValue(map[string]any{})
				if err != nil {
					return nil, err
				}

				return ast.NewTerm(astVal), nil
			case err != nil:
				return nil, err
			}

			buf := new(bytes.Buffer)

			var result proto.Message

			if resp != nil {
				result = resp
			}

			if err := builtins.ProtoToBuf(buf, result); err != nil {
				return nil, err
			}

			v, err := ast.ValueFromReader(buf)
			if err != nil {
				return nil, err
			}

			return ast.NewTerm(v), nil
		}
}
