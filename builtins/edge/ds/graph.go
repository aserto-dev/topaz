package ds

import (
	"bytes"

	v2 "github.com/aserto-dev/go-directory/aserto/directory/common/v2"
	ds2 "github.com/aserto-dev/go-directory/aserto/directory/reader/v2"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// RegisterGraph - ds.graph
//
// ds.graph({
//     "id": "",
//     "object_id": "",
//     "object_type": "",
//     "relation": "",
//     "subject_id": "",
//     "subject_type": ""
//   })
//
func RegisterGraph(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: false,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {

			type args struct {
				AnchorID    string `json:"id"`
				SubjectType string `json:"subject_type"`
				SubjectID   string `json:"subject_id"`
				Relation    string `json:"relation"`
				ObjectType  string `json:"object_type"`
				ObjectID    string `json:"object_id"`
			}

			var a args
			if err := ast.As(op1.Value, &a); err != nil {
				return nil, err
			}

			client, err := dr.GetDS(bctx.Context)
			if err != nil {
				return nil, errors.Wrapf(err, "get directory client")
			}

			resp, err := client.GetGraph(bctx.Context, &ds2.GetGraphRequest{
				Anchor: &v2.ObjectIdentifier{
					Id: &a.AnchorID,
				},
				Subject: &v2.ObjectIdentifier{
					Type: &a.SubjectType,
					Id:   &a.SubjectID,
				},
				Relation: &v2.RelationTypeIdentifier{
					Name: &a.Relation,
				},
				Object: &v2.ObjectIdentifier{
					Id:   &a.ObjectID,
					Type: &a.ObjectType,
				},
			})
			if err != nil {
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
