package ds

import (
	"bytes"

	dsc2 "github.com/aserto-dev/go-directory/aserto/directory/common/v2"
	dsr2 "github.com/aserto-dev/go-directory/aserto/directory/reader/v2"
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

			type args struct {
				Anchor   *dsc2.ObjectIdentifier       `json:"anchor"`
				Subject  *dsc2.ObjectIdentifier       `json:"subject"`
				Relation *dsc2.RelationTypeIdentifier `json:"relation"`
				Object   *dsc2.ObjectIdentifier       `json:"object"`
			}

			var a args
			if err := ast.As(op1.Value, &a); err != nil {
				return nil, err
			}

			if a.Anchor == nil && a.Subject == nil && a.Relation == nil && a.Object == nil {
				a = args{
					Anchor: &dsc2.ObjectIdentifier{
						Type: proto.String(""),
						Key:  proto.String(""),
					},
					Subject: &dsc2.ObjectIdentifier{
						Type: proto.String(""),
						Key:  proto.String(""),
					},
					Relation: &dsc2.RelationTypeIdentifier{
						Name: proto.String(""),
					},
					Object: &dsc2.ObjectIdentifier{
						Type: proto.String(""),
						Key:  proto.String(""),
					},
				}
				return help(fnName, a)
			}

			client, err := dr.GetDS(bctx.Context)
			if err != nil {
				return nil, errors.Wrapf(err, "get directory client")
			}

			resp, err := client.GetGraph(bctx.Context, &dsr2.GetGraphRequest{
				Anchor:   a.Anchor,
				Subject:  a.Subject,
				Relation: a.Relation,
				Object:   a.Object,
			})
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
