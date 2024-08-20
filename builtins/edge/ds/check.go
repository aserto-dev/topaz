package ds

import (
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/resolvers"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"

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
func RegisterCheck(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
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

			resp, err := dr.GetDS().Check(bctx.Context, &args)
			if err != nil {
				traceError(&bctx, fnName, err)
				return nil, err
			}

			return ast.BooleanTerm(resp.Check), nil
		}
}

// RegisterCheckRelation - ds.check_relation
//
//	ds.check_relation: {
//		"object_id": "",
//		"object_type": "",
//		"relation": "",
//		"subject_id": "",
//		"subject_type": "",
//		"trace": false
//	  }
func RegisterCheckRelation(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.B),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args dsr3.CheckRelationRequest

			if err := ast.As(op1.Value, &args); err != nil {
				traceError(&bctx, fnName, err)
				return nil, err
			}

			if proto.Equal(&args, &dsr3.CheckRelationRequest{}) {
				return helpMsg(fnName, &dsr3.CheckRelationRequest{
					ObjectType:  "",
					ObjectId:    "",
					Relation:    "",
					SubjectType: "",
					SubjectId:   "",
					Trace:       false,
				})
			}

			//nolint: staticcheck // SA1019: client.CheckRelation is deprecated
			resp, err := dr.GetDS().CheckRelation(bctx.Context, &args)
			if err != nil {
				traceError(&bctx, fnName, err)
				return nil, err
			}

			return ast.BooleanTerm(resp.Check), nil
		}
}

// RegisterCheckPermission - ds.check_permission
//
//	ds.check_permission: {
//		"object_id": "",
//		"object_type": "",
//		"permission": "",
//		"subject_id": "",
//		"subject_type": "",
//		"trace": false
//	  }
func RegisterCheckPermission(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.B),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args dsr3.CheckPermissionRequest

			if err := ast.As(op1.Value, &args); err != nil {
				traceError(&bctx, fnName, err)
				return nil, err
			}

			if proto.Equal(&args, &dsr3.CheckPermissionRequest{}) {
				return helpMsg(fnName, &dsr3.CheckPermissionRequest{
					ObjectType:  "",
					ObjectId:    "",
					Permission:  "",
					SubjectType: "",
					SubjectId:   "",
					Trace:       false,
				})
			}

			//nolint: staticcheck // SA1019: client.CheckPermission is deprecated
			resp, err := dr.GetDS().CheckPermission(bctx.Context, &args)
			if err != nil {
				traceError(&bctx, fnName, err)
				return nil, err
			}

			return ast.BooleanTerm(resp.Check), nil
		}
}
