package ds

import (
	v2 "github.com/aserto-dev/go-directory/aserto/directory/common/v2"
	ds2 "github.com/aserto-dev/go-directory/aserto/directory/reader/v2"
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
				Subject      *v2.ObjectIdentifier       `json:"subject"`
				RelationType *v2.RelationTypeIdentifier `json:"relation"`
				Object       *v2.ObjectIdentifier       `json:"object"`
			}

			var a args

			if err := ast.As(op1.Value, &a); err != nil {
				return nil, err
			}

			client, err := dr.GetDS(bctx.Context)
			if err != nil {
				return nil, errors.Wrapf(err, "get directory client")
			}

			/* TODO: Enable subject validation
			if !ValidateObject(a.Subject){
				return nil, errors.Errorf("invalid subject arguments")
			}
			*/

			if !ValidateObject(a.Object) {
				return nil, errors.Errorf("invalid object arguments")
			}

			if !ValidateRelationType(a.RelationType) {
				return nil, errors.Errorf("invalid relation arguments")
			}

			resp, err := client.CheckRelation(bctx.Context, &ds2.CheckRelationRequest{
				Subject:  a.Subject,
				Relation: a.RelationType,
				Object:   a.Object,
				Trace:    false,
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
				Subject    *v2.ObjectIdentifier     `json:"subject"`
				Permission *v2.PermissionIdentifier `json:"permission"`
				Object     *v2.ObjectIdentifier     `json:"object"`
			}

			var a args

			if err := ast.As(op1.Value, &a); err != nil {
				return nil, err
			}

			if (args{}) == a {
				return help(fnName, args{})
			}

			/* TODO: Enable subject validation
			if !ValidateObject(a.Subject) {
				return nil, errors.Errorf("invalid subject arguments")
			}
			*/
			if !ValidateObject(a.Object) {
				return nil, errors.Errorf("invalid object arguments")
			}

			if !ValidatePermissionType(a.Permission) {
				return nil, errors.Errorf("invalid permission arguments")
			}

			client, err := dr.GetDS(bctx.Context)
			if err != nil {
				return nil, errors.Wrapf(err, "get directory client")
			}

			resp, err := client.CheckPermission(bctx.Context, &ds2.CheckPermissionRequest{
				Subject: a.Subject,

				Permission: a.Permission,

				Object: a.Object,
				Trace:  false,
			})
			if err != nil {
				return nil, err
			}

			return ast.BooleanTerm(resp.Check), nil
		}
}
