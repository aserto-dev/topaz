package ds

import (
	"bytes"

	dsc "github.com/aserto-dev/go-directory/aserto/directory/common/v2"
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v2"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"

	"github.com/rs/zerolog"
)

// RegisterRelation - ds.relation
//
//	ds.relation({
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
//		},
//		"with_objects": false
//	  })
type extendedRelation struct {
	*dsc.RelationIdentifier
	WithObjects bool `json:"with_objects"`
}

func RegisterRelation(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var a *extendedRelation
			if err := ast.As(op1.Value, &a); err != nil {
				return nil, err
			}

			if a == nil || a.RelationIdentifier == nil {

				a = &extendedRelation{
					RelationIdentifier: &dsc.RelationIdentifier{
						Subject: &dsc.ObjectIdentifier{
							Type: proto.String(""),
							Key:  proto.String(""),
						},
						Relation: &dsc.RelationTypeIdentifier{
							Name: proto.String(""),
						},
						Object: &dsc.ObjectIdentifier{
							Type: proto.String(""),
							Key:  proto.String(""),
						},
					},
					WithObjects: false,
				}
				return help(fnName, a)
			}

			client, err := dr.GetDS(bctx.Context)
			if err != nil {
				return nil, errors.Wrapf(err, "get directory client")
			}

			resp, err := client.GetRelation(bctx.Context, &reader.GetRelationRequest{Param: a.RelationIdentifier, WithObjects: &a.WithObjects})
			if err != nil {
				traceError(&bctx, fnName, err)
				return nil, err
			}

			buf := new(bytes.Buffer)
			if resp != nil {
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
