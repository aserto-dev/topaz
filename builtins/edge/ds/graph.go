package ds

import (
	"bytes"

	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/resolvers"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"

	"github.com/rs/zerolog"
	"google.golang.org/protobuf/proto"
)

// RegisterGraph - ds.graph
//
//	ds.graph({
//	    "object_type": "",
//	    "object_id": "",
//	    "relation": "",
//	    "subject_type": "",
//	    "subject_id": "",
//	    "subject_relation": "",
//	    "explain": false,
//	    "trace": false
//	}
func RegisterGraph(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args dsr3.GetGraphRequest

			if err := ast.As(op1.Value, &args); err != nil {
				traceError(&bctx, fnName, err)
				return nil, err
			}

			if proto.Equal(&args, &dsr3.GetGraphRequest{}) {
				return helpMsg(fnName, &dsr3.GetGraphRequest{
					ObjectType:      "",
					ObjectId:        "",
					Relation:        "",
					SubjectType:     "",
					SubjectId:       "",
					SubjectRelation: "",
					Explain:         false,
					Trace:           false,
				})
			}

			resp, err := dr.GetDS().GetGraph(bctx.Context, &args)
			if err != nil {
				traceError(&bctx, fnName, err)
				return nil, err
			}

			buf := new(bytes.Buffer)
			if len(resp.Results) > 0 {
				if err := ProtoToBuf(buf, resp); err != nil {
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
