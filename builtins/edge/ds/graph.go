package ds

import (
	"bytes"

	dsc2 "github.com/aserto-dev/go-directory/aserto/directory/common/v2"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/resolvers"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/proto"
)

// RegisterGraph - ds.graph
//
//	v3 (latest) request format:
//
//	ds.graph({
//	    "anchor_type": "",
//	    "anchor_id": "",
//	    "object_type": "",
//	    "object_id": "",
//	    "relation": "",
//	    "subject_type": "",
//	    "subject_id": ""
//	    "subject_relation": ""
//	}
//
//	v2 request format:
//
//	ds.graph({
//		"anchor": {
//		  "type": ""
//		  "key": "",
//		},
//		"object": {
//		  "type": ""
//		  "key": "",
//		},
//		"relation": {
//		  "name": "",
//		},
//		"subject": {
//		  "type": ""
//		  "key": "",
//		}
//	})
func RegisterGraph(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {

			var args dsr3.GetGraphRequest

			if err := ast.As(op1.Value, &args); err != nil {

				// if v3 input parsing fails, fallback to v2 before exiting with an error.
				type argsV2 struct {
					Anchor   *dsc2.ObjectIdentifier       `json:"anchor"`
					Subject  *dsc2.ObjectIdentifier       `json:"subject"`
					Relation *dsc2.RelationTypeIdentifier `json:"relation"`
					Object   *dsc2.ObjectIdentifier       `json:"object"`
				}

				var a2 argsV2
				if err := ast.As(op1.Value, &a2); err != nil {
					return nil, err
				}

				args = dsr3.GetGraphRequest{
					AnchorType:  a2.Anchor.GetType(),
					AnchorId:    a2.Anchor.GetKey(),
					ObjectType:  a2.Object.GetType(),
					ObjectId:    a2.Object.GetKey(),
					Relation:    a2.Relation.GetName(),
					SubjectType: a2.Subject.GetType(),
					SubjectId:   a2.Subject.GetKey(),
				}
			}

			if proto.Equal(&args, &dsr3.GetGraphRequest{}) {
				return helpMsg(fnName, &dsr3.GetGraphRequest{
					AnchorType:      "",
					AnchorId:        "",
					ObjectType:      "",
					ObjectId:        "",
					Relation:        "",
					SubjectType:     "",
					SubjectId:       "",
					SubjectRelation: "",
				})
			}

			client, err := dr.GetDS(bctx.Context)
			if err != nil {
				return nil, errors.Wrapf(err, "get directory client")
			}

			resp, err := client.GetGraph(bctx.Context, &args)
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
