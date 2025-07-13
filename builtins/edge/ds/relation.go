package ds

import (
	"bytes"

	"github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/builtins"
	"github.com/aserto-dev/topaz/pkg/cli/x"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/samber/lo"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/types"

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
			var args reader.GetRelationRequest

			if err := ast.As(op1.Value, &args); err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			if proto.Equal(&args, &reader.GetRelationRequest{}) {
				return builtins.HelpMsg(fnName, getRelationReq())
			}

			resp, err := dr.GetDS().GetRelation(bctx.Context, &args)

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

func getRelationReq() *reader.GetRelationRequest {
	return &reader.GetRelationRequest{
		ObjectType:      "",
		ObjectId:        "",
		Relation:        "",
		SubjectType:     "",
		SubjectId:       "",
		SubjectRelation: "",
		WithObjects:     false,
	}
}

// RegisterRelations - ds.relations
//
//	ds.relations: {
//		object_type: "",
//		object_id: "",
//		relation: "",
//		subject_type: "",
//		subject_id: "",
//		subject_relation: "",
//		with_objects: false,
//		with_empty_subject_relation: false
//	}
func RegisterRelations(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args reader.GetRelationsRequest

			if err := ast.As(op1.Value, &args); err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			if proto.Equal(&args, &reader.GetRelationsRequest{}) {
				return builtins.HelpMsg(fnName, getRelationsReq())
			}

			args.Page = &common.PaginationRequest{Size: x.MaxPaginationSize, Token: ""}

			resp := &reader.GetRelationsResponse{}

			for {
				r, err := dr.GetDS().GetRelations(bctx.Context, &args)
				if err != nil {
					builtins.TraceError(&bctx, fnName, err)
					return nil, err
				}

				resp.Results = append(resp.GetResults(), r.GetResults()...)
				resp.Objects = lo.Assign(resp.GetObjects(), r.GetObjects())

				if r.GetPage().GetNextToken() == "" {
					break
				}

				args.Page.Token = r.GetPage().GetNextToken()
			}

			buf := new(bytes.Buffer)

			var result proto.Message
			if resp.Results != nil {
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

func getRelationsReq() *reader.GetRelationsRequest {
	return &reader.GetRelationsRequest{
		ObjectType:               "",
		ObjectId:                 "",
		Relation:                 "",
		SubjectType:              "",
		SubjectId:                "",
		SubjectRelation:          "",
		WithObjects:              false,
		WithEmptySubjectRelation: false,
	}
}
