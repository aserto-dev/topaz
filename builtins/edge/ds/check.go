package ds

import (
	ds2 "github.com/aserto-dev/go-directory/aserto/directory/v2"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// RegisterCheckRelation - ds.check_relation
//
// ds.relation({
//     "object": {
//       "id": "",
//       "key": "",
//       "type": ""
//     },
//     "relation": {
//       "name": "",
//       "object_type": ""
//     },
//     "subject": {
//       "id": "",
//       "key": "",
//       "type": ""
//     }
//   })
//
func RegisterCheckRelation(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.B),
			Memoize: false,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {

			type args struct {
				Subject      ObjectParam       `json:"subject"`
				RelationType RelationTypeParam `json:"relation"`
				Object       ObjectParam       `json:"object"`
			}

			var a args

			if err := ast.As(op1.Value, &a); err != nil {
				return nil, err
			}

			if (args{}) == a {
				return help(fnName, args{})
			}

			client, err := dr.GetDS(bctx.Context)
			if err != nil {
				return nil, errors.Wrapf(err, "get directory client")
			}

			subjectParam := a.Subject.Validate()
			if subjectParam == nil {
				return nil, errors.Errorf("invalid subject arguments")
			}

			objectParam := a.Object.Validate()
			if objectParam == nil {
				return nil, errors.Errorf("invalid object arguments")
			}

			relationTypeParam := a.RelationType.Validate()
			if relationTypeParam == nil {
				return nil, errors.Errorf("invalid relation arguments")
			}

			resp, err := client.Check(bctx.Context, &ds2.CheckRequest{
				Subject: subjectParam,
				Check: &ds2.CheckRequest_Relation{
					Relation: relationTypeParam,
				},
				Object: objectParam,
				Trace:  false,
			})
			if err != nil {
				return nil, err
			}

			return ast.BooleanTerm(resp.Check), nil
		}
}

// RegisterCheckPermission - ds.check_permission
//
// ds.check_permission({
//     "object": {
//       "id": "",
//       "key": "",
//       "type": ""
//     },
//     "permission": {
//       "id": "",
//       "name": ""
//     },
//     "subject": {
//       "id": "",
//       "key": "",
//       "type": ""
//     }
//   })
//
func RegisterCheckPermission(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.B),
			Memoize: false,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			type args struct {
				Subject        ObjectParam         `json:"subject"`
				PermissionType PermissionTypeParam `json:"permission"`
				Object         ObjectParam         `json:"object"`
			}

			var a args

			if err := ast.As(op1.Value, &a); err != nil {
				return nil, err
			}

			if (args{}) == a {
				return help(fnName, args{})
			}

			subjectParam := a.Subject.Validate()
			if subjectParam == nil {
				return nil, errors.Errorf("invalid subject arguments")
			}

			objectParam := a.Object.Validate()
			if objectParam == nil {
				return nil, errors.Errorf("invalid object arguments")
			}

			permissionTypeParam := a.PermissionType.Validate()
			if permissionTypeParam == nil {
				return nil, errors.Errorf("invalid permission arguments")
			}

			client, err := dr.GetDS(bctx.Context)
			if err != nil {
				return nil, errors.Wrapf(err, "get directory client")
			}

			resp, err := client.Check(bctx.Context, &ds2.CheckRequest{
				Subject: subjectParam,
				Check: &ds2.CheckRequest_Permission{
					Permission: permissionTypeParam,
				},
				Object: objectParam,
				Trace:  false,
			})
			if err != nil {
				return nil, err
			}

			return ast.BooleanTerm(resp.Check), nil
		}
}
