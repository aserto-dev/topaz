package ds

import (
	"bytes"

	dsc2 "github.com/aserto-dev/go-directory/aserto/directory/common/v2"
	dsc3 "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsr2 "github.com/aserto-dev/go-directory/aserto/directory/reader/v2"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/go-edge-ds/pkg/convert"
	"github.com/aserto-dev/topaz/resolvers"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/proto"
)

// RegisterRelation - ds.relation
//
// v3 (latest) request format:
//
//	ds.relation: {
//		"object_id": "",
//		"object_type": "",
//		"relation": "",
//		"subject_id": "",
//		"subject_relation": "",
//		"subject_type": "",
//		"with_objects": false
//	  }
//
// v2 request format:
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
	*dsc2.RelationIdentifier
	WithObjects bool `json:"with_objects"`
}

func RegisterRelation(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {

			var (
				args     dsr3.GetRelationRequest
				outputV2 bool
			)

			if err := ast.As(op1.Value, &args); err != nil {

				// if v3 input parsing fails, fallback to v2 before exiting with an error.
				var a2 extendedRelation
				if err := ast.As(op1.Value, &a2); err != nil {
					return nil, err
				}

				outputV2 = true

				args = dsr3.GetRelationRequest{
					ObjectType:      a2.GetObject().GetType(),
					ObjectId:        a2.GetObject().GetKey(),
					Relation:        a2.GetRelation().GetName(),
					SubjectType:     a2.GetSubject().GetType(),
					SubjectId:       a2.GetSubject().GetKey(),
					SubjectRelation: "",
					WithObjects:     a2.WithObjects,
				}
			}

			if proto.Equal(&args, &dsr3.GetRelationRequest{}) {
				return helpMsg(fnName, &dsr3.GetRelationRequest{
					ObjectType:      "",
					ObjectId:        "",
					Relation:        "",
					SubjectType:     "",
					SubjectId:       "",
					SubjectRelation: "",
					WithObjects:     false,
				})
			}

			client, err := dr.GetDS(bctx.Context)
			if err != nil {
				return nil, errors.Wrapf(err, "get directory client")
			}

			resp, err := client.GetRelation(bctx.Context, &args)
			if err != nil {
				traceError(&bctx, fnName, err)
				return nil, err
			}

			buf := new(bytes.Buffer)
			var result proto.Message

			if resp != nil {
				result = resp
				if outputV2 {
					result = &dsr2.GetRelationResponse{
						Results: []*dsc2.Relation{convert.RelationToV2(resp.Result)},
						Objects: convert.ObjectMapToV2(resp.Objects),
					}
				}

				if err := ProtoToBuf(buf, result); err != nil {
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

// RegisterRelations - ds.relations
//
// v3 (latest) request format:
//
// v2 request format:
//
//	ds.relations({
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
//	  })
func RegisterRelations(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {

			var (
				args     dsr3.GetRelationsRequest
				outputV2 bool
			)

			if err := ast.As(op1.Value, &args); err != nil {
				// if v3 input parsing fails, fallback to v2 before exiting with an error.
				var a2 dsc2.RelationIdentifier
				if err := ast.As(op1.Value, &a2); err != nil {
					return nil, err
				}

				outputV2 = true

				args = dsr3.GetRelationsRequest{
					ObjectType:      a2.GetObject().GetType(),
					ObjectId:        a2.GetObject().GetKey(),
					Relation:        a2.GetRelation().GetName(),
					SubjectType:     a2.GetSubject().GetType(),
					SubjectId:       a2.GetSubject().GetKey(),
					SubjectRelation: "",
					WithObjects:     false,
				}

			}

			if proto.Equal(&args, &dsr3.GetRelationsRequest{}) {
				return helpMsg(fnName, &dsr3.GetRelationsRequest{})
			}

			client, err := dr.GetDS(bctx.Context)
			if err != nil {
				return nil, errors.Wrapf(err, "get directory client")
			}

			args.Page = &dsc3.PaginationRequest{Size: 100, Token: ""}

			resp := &dsr3.GetRelationsResponse{}

			for {
				r, err := client.GetRelations(bctx.Context, &args)
				if err != nil {
					traceError(&bctx, fnName, err)
					return nil, err
				}

				resp.Results = append(resp.Results, r.Results...)

				if r.Page.NextToken == "" {
					break
				}
				args.Page.Token = r.Page.NextToken
			}

			buf := new(bytes.Buffer)
			var result proto.Message

			if resp.Results != nil {
				result = resp
				if outputV2 {
					result = &dsr2.GetRelationsResponse{
						Results: convert.RelationArrayToV2(resp.Results),
					}
				}

				if err := ProtoToBuf(buf, result); err != nil {
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
