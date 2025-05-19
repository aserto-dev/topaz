package ds

import (
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/types"

	"github.com/rs/zerolog"
	"google.golang.org/protobuf/proto"
)

// RegisterCheck - ds.check
//
//	ds.check({
//	  "object_type": "",
//	  "object_id": "",
//	  "relation": "",
//	  "subject_type": ""
//	  "subject_id": "",
//	  "trace": false
//	})
func RegisterCheck(logger *zerolog.Logger, fnName string, dr dsr3.ReaderClient) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.B),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args dsr3.CheckRequest

			if err := ast.As(op1.Value, &args); err != nil {
				return nil, err
			}

			if proto.Equal(&args, &dsr3.CheckRequest{}) {
				return helpMsg(fnName, &dsr3.CheckRequest{
					ObjectType:  "",
					ObjectId:    "",
					Relation:    "",
					SubjectType: "",
					SubjectId:   "",
				})
			}

			resp, err := dr.Check(bctx.Context, &args)
			if err != nil {
				traceError(&bctx, fnName, err)
				return nil, err
			}

			return ast.BooleanTerm(resp.GetCheck()), nil
		}
}
