package ds

import (
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

// RegisterCheck - ds.check
//
// v3 (latest) request format:
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

			client, err := dr.GetDS(bctx.Context)
			if err != nil {
				return nil, errors.Wrapf(err, "get directory client")
			}

			resp, err := client.Check(bctx.Context, &args)
			if err != nil {
				traceError(&bctx, fnName, err)
				return nil, err
			}

			return ast.BooleanTerm(resp.Check), nil
		}
}

// RegisterCheckRelation - ds.check_relation
//
// v3 (latest) request format:
//
//	ds.check_relation: {
//		"object_id": "",
//		"object_type": "",
//		"relation": "",
//		"subject_id": "",
//		"subject_type": "",
//		"trace": false
//	  }
//
// v2 request format:
//
//	ds.check_relation({
//	  "object": {
//	    "type": ""
//	    "key": "",
//	  },
//	  "relation": {
//	    "name": "",
//	  },
//	  "subject": {
//	    "type": ""
//	    "key": "",
//	  }
//	})
func RegisterCheckRelation(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.B),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {

			var args dsr3.CheckRelationRequest

			if err := ast.As(op1.Value, &args); err != nil {

				// if v3 input parsing fails, fallback to v2 before exiting with an error.
				type argsV2 struct {
					Subject      *dsc2.ObjectIdentifier       `json:"subject"`
					RelationType *dsc2.RelationTypeIdentifier `json:"relation"`
					Object       *dsc2.ObjectIdentifier       `json:"object"`
				}

				var a2 argsV2
				if err := ast.As(op1.Value, &a2); err != nil {
					return nil, err
				}

				if a2.RelationType.GetObjectType() == "" {
					a2.RelationType.ObjectType = a2.Object.Type
				}

				args = dsr3.CheckRelationRequest{
					ObjectType:  a2.Object.GetType(),
					ObjectId:    a2.Object.GetKey(),
					Relation:    a2.RelationType.GetName(),
					SubjectType: a2.Subject.GetType(),
					SubjectId:   a2.Subject.GetKey(),
					Trace:       false,
				}
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

			client, err := dr.GetDS(bctx.Context)
			if err != nil {
				return nil, errors.Wrapf(err, "get directory client")
			}

			resp, err := client.CheckRelation(bctx.Context, &args)
			if err != nil {
				traceError(&bctx, fnName, err)
				return nil, err
			}

			return ast.BooleanTerm(resp.Check), nil
		}
}

// RegisterCheckPermission - ds.check_permission
//
// v3 (latest) request format:
//
//	ds.check_permission: {
//		"object_id": "",
//		"object_type": "",
//		"permission": "",
//		"subject_id": "",
//		"subject_type": "",
//		"trace": false
//	  }
//
// v2 request format:
//
//	ds.check_permission({
//		"object": {
//		  "type": ""
//		  "key": "",
//		},
//		"permission": {
//		  "name": ""
//		},
//		"subject": {
//		  "type": ""
//		  "key": "",
//		}
//	})
func RegisterCheckPermission(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.B),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {

			var args dsr3.CheckPermissionRequest

			if err := ast.As(op1.Value, &args); err != nil {

				// if v3 input parsing fails, fallback to v2 before exiting with an error.
				type argsV2 struct {
					Subject    *dsc2.ObjectIdentifier     `json:"subject"`
					Permission *dsc2.PermissionIdentifier `json:"permission"`
					Object     *dsc2.ObjectIdentifier     `json:"object"`
				}

				var a2 argsV2
				if err := ast.As(op1.Value, &a2); err != nil {
					return nil, err
				}

				args = dsr3.CheckPermissionRequest{
					ObjectType:  a2.Object.GetType(),
					ObjectId:    a2.Object.GetKey(),
					Permission:  a2.Permission.GetName(),
					SubjectType: a2.Subject.GetType(),
					SubjectId:   a2.Subject.GetKey(),
					Trace:       false,
				}
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

			client, err := dr.GetDS(bctx.Context)
			if err != nil {
				return nil, errors.Wrapf(err, "get directory client")
			}

			resp, err := client.CheckPermission(bctx.Context, &args)
			if err != nil {
				traceError(&bctx, fnName, err)
				return nil, err
			}

			return ast.BooleanTerm(resp.Check), nil
		}
}
