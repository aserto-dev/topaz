package ds

import (
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

// RegisterCheckRelation - ds.check_relation
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

			type args struct {
				Subject      *dsc2.ObjectIdentifier       `json:"subject"`
				RelationType *dsc2.RelationTypeIdentifier `json:"relation"`
				Object       *dsc2.ObjectIdentifier       `json:"object"`
			}

			var a args

			if err := ast.As(op1.Value, &a); err != nil {
				return nil, err
			}

			if a.Subject == nil && a.RelationType == nil && a.Object == nil {
				a = args{
					Subject: &dsc2.ObjectIdentifier{
						Type: proto.String(""),
						Key:  proto.String(""),
					},
					RelationType: &dsc2.RelationTypeIdentifier{
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

			if a.RelationType.GetObjectType() == "" {
				a.RelationType.ObjectType = a.Object.Type
			}

			resp, err := client.CheckRelation(bctx.Context, &dsr2.CheckRelationRequest{
				Subject:  a.Subject,
				Relation: a.RelationType,
				Object:   a.Object,
				Trace:    false,
			})
			if err != nil {
				traceError(&bctx, fnName, err)
				return nil, err
			}

			return ast.BooleanTerm(resp.Check), nil
		}
}

// RegisterCheckPermission - ds.check_permission
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

			type args struct {
				Subject    *dsc2.ObjectIdentifier     `json:"subject"`
				Permission *dsc2.PermissionIdentifier `json:"permission"`
				Object     *dsc2.ObjectIdentifier     `json:"object"`
			}

			var a args

			if err := ast.As(op1.Value, &a); err != nil {
				return nil, err
			}

			if a.Subject == nil && a.Permission == nil && a.Object == nil {
				a = args{
					Subject: &dsc2.ObjectIdentifier{
						Type: proto.String(""),
						Key:  proto.String(""),
					},
					Permission: &dsc2.PermissionIdentifier{
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

			resp, err := client.CheckPermission(bctx.Context, &dsr2.CheckPermissionRequest{
				Subject: a.Subject,

				Permission: a.Permission,

				Object: a.Object,
				Trace:  false,
			})
			if err != nil {
				traceError(&bctx, fnName, err)
				return nil, err
			}

			return ast.BooleanTerm(resp.Check), nil
		}
}
