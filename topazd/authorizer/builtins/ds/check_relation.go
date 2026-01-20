package ds

import (
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/topazd/authorizer/builtins"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/types"

	"github.com/rs/zerolog"
	"google.golang.org/protobuf/proto"
)

const dsCheckRelationHelp string = `ds.check_relation({
	"object_id": "",
	"object_type": "",
	"relation": "",
	"subject_id": "",
	"subject_type": "",
	"trace": false
})`

// RegisterCheckRelation - ds.check_relation (OBSOLETE).
func RegisterCheckRelation(logger *zerolog.Logger, fnName string, dr reader.ReaderClient) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.B),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args reader.CheckRelationRequest

			if err := ast.As(op1.Value, &args); err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			if proto.Equal(&args, &reader.CheckRelationRequest{}) {
				return ast.StringTerm(dsCheckRelationHelp), nil
			}

			//nolint: staticcheck // SA1019: client.CheckRelation is deprecated
			resp, err := dr.CheckRelation(bctx.Context, &args)
			if err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			return ast.BooleanTerm(resp.GetCheck()), nil
		}
}
