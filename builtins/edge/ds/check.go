package ds

import (
	dsc "github.com/aserto-dev/go-directory/aserto/directory/common/v2"
	dsr "github.com/aserto-dev/go-directory/aserto/directory/reader/v2"
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
//	    "key": "",
//	    "type": ""
//	  },
//	  "relation": {
//	    "name": "",
//	    "object_type": ""
//	  },
//	  "subject": {
//	    "key": "",
//	    "type": ""
//	  }
//	})
func RegisterCheckRelation(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.B),
			Memoize: false,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {

			type args struct {
				Subject      *dsc.ObjectIdentifier       `json:"subject"`
				RelationType *dsc.RelationTypeIdentifier `json:"relation"`
				Object       *dsc.ObjectIdentifier       `json:"object"`
			}

			var a args

			if err := ast.As(op1.Value, &a); err != nil {
				return nil, err
			}

			if a.Subject == nil && a.RelationType == nil && a.Object == nil {
				a = args{
					Subject: &dsc.ObjectIdentifier{
						Type: proto.String(""),
						Key:  proto.String(""),
					},
					RelationType: &dsc.RelationTypeIdentifier{
						ObjectType: proto.String(""),
						Name:       proto.String(""),
					},
					Object: &dsc.ObjectIdentifier{
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

			resp, err := client.CheckRelation(bctx.Context, &dsr.CheckRelationRequest{
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
//		  "key": "",
//		  "type": ""
//		},
//		"permission": {
//		  "name": ""
//		},
//		"subject": {
//		  "key": "",
//		  "type": ""
//		}
//	})
func RegisterCheckPermission(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.B),
			Memoize: false,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {

			type args struct {
				Subject    *dsc.ObjectIdentifier     `json:"subject"`
				Permission *dsc.PermissionIdentifier `json:"permission"`
				Object     *dsc.ObjectIdentifier     `json:"object"`
			}

			var a args

			if err := ast.As(op1.Value, &a); err != nil {
				return nil, err
			}

			if a.Subject == nil && a.Permission == nil && a.Object == nil {
				a = args{
					Subject: &dsc.ObjectIdentifier{
						Type: proto.String(""),
						Key:  proto.String(""),
					},
					Permission: &dsc.PermissionIdentifier{
						Name: proto.String(""),
					},
					Object: &dsc.ObjectIdentifier{
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

			resp, err := client.CheckPermission(bctx.Context, &dsr.CheckPermissionRequest{
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
