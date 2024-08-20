package ds

import (
	"bytes"

	dsc3 "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/resolvers"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"

	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// RegisterRelation - ds.relation
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
func RegisterRelation(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args dsr3.GetRelationRequest

			if err := ast.As(op1.Value, &args); err != nil {
				traceError(&bctx, fnName, err)
				return nil, err
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

			resp, err := dr.GetDS().GetRelation(bctx.Context, &args)
			switch {
			case status.Code(err) == codes.NotFound:
				traceError(&bctx, fnName, err)
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

			if err := ProtoToBuf(buf, result); err != nil {
				return nil, err
			}

			v, err := ast.ValueFromReader(buf)
			if err != nil {
				return nil, err
			}

			return ast.NewTerm(v), nil
		}
}

// TODO : update v3 format
//
//	RegisterRelations - ds.relations({
//		"ds.relations": {
//			"object_type": "",
//			"object_id": "",
//			"relation": "",
//			"subject_type": "",
//			"subject_id": "",
//			"subject_relation": "",
//			"with_objects": false,
//			"with_empty_subject_relation": false
//		}
//	})
func RegisterRelations(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args dsr3.GetRelationsRequest

			if err := ast.As(op1.Value, &args); err != nil {
				traceError(&bctx, fnName, err)
				return nil, err
			}

			if proto.Equal(&args, &dsr3.GetRelationsRequest{}) {
				return helpMsg(fnName, &dsr3.GetRelationsRequest{})
			}

			args.Page = &dsc3.PaginationRequest{Size: 100, Token: ""}

			resp := &dsr3.GetRelationsResponse{}

			for {
				r, err := dr.GetDS().GetRelations(bctx.Context, &args)
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
			}

			if err := ProtoToBuf(buf, result); err != nil {
				return nil, err
			}

			v, err := ast.ValueFromReader(buf)
			if err != nil {
				return nil, err
			}

			return ast.NewTerm(v), nil
		}
}
