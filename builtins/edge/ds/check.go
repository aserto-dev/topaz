package ds

import (
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/builtins"
	"github.com/aserto-dev/topaz/resolvers"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/types"

	"github.com/rs/zerolog"
	"google.golang.org/protobuf/proto"
)

/*
	ds.check({
	  "object_type": "",
	  "object_id": "",
	  "relation": "",
	  "subject_type": ""
	  "subject_id": "",
	  "trace": false
	})
*/
func RegisterCheck(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.B),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args reader.CheckRequest

			if err := ast.As(op1.Value, &args); err != nil {
				return nil, err
			}

			if proto.Equal(&args, &reader.CheckRequest{}) {
				return builtins.HelpMsg(fnName, checkReq())
			}

			resp, err := dr.GetDS().Check(bctx.Context, &args)
			if err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			return ast.BooleanTerm(resp.GetCheck()), nil
		}
}

func checkReq() *reader.CheckRequest {
	return &reader.CheckRequest{
		ObjectType:  "",
		ObjectId:    "",
		Relation:    "",
		SubjectType: "",
		SubjectId:   "",
	}
}

/*
	ds.check_relation: {
		"object_id": "",
		"object_type": "",
		"relation": "",
		"subject_id": "",
		"subject_type": "",
		"trace": false
	}
*/

func RegisterCheckRelation(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.B),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args reader.CheckRelationRequest

			if err := ast.As(op1.Value, &args); err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			if proto.Equal(&args, &reader.CheckRelationRequest{}) {
				return builtins.HelpMsg(fnName, checkRelationReq())
			}

			//nolint: staticcheck // SA1019: client.CheckRelation is deprecated
			resp, err := dr.GetDS().CheckRelation(bctx.Context, &args)
			if err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			return ast.BooleanTerm(resp.GetCheck()), nil
		}
}

func checkRelationReq() *reader.CheckRelationRequest {
	return &reader.CheckRelationRequest{
		ObjectType:  "",
		ObjectId:    "",
		Relation:    "",
		SubjectType: "",
		SubjectId:   "",
		Trace:       false,
	}
}

/*
	ds.check_permission: {
		"object_id": "",
		"object_type": "",
		"permission": "",
		"subject_id": "",
		"subject_type": "",
		"trace": false
	}
*/

func RegisterCheckPermission(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.B),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args reader.CheckPermissionRequest

			if err := ast.As(op1.Value, &args); err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			if proto.Equal(&args, &reader.CheckPermissionRequest{}) {
				return builtins.HelpMsg(fnName, checkPermissionReq())
			}

			//nolint: staticcheck // SA1019: client.CheckPermission is deprecated
			resp, err := dr.GetDS().CheckPermission(bctx.Context, &args)
			if err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			return ast.BooleanTerm(resp.GetCheck()), nil
		}
}

func checkPermissionReq() *reader.CheckPermissionRequest {
	return &reader.CheckPermissionRequest{
		ObjectType:  "",
		ObjectId:    "",
		Permission:  "",
		SubjectType: "",
		SubjectId:   "",
		Trace:       false,
	}
}
