package ds

import (
	"bytes"

	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/resolvers"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"

	"github.com/rs/zerolog"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// RegisterCheck - ds.checks
//
//	ds.checks({
//	  "object_type": "",
//	  "object_id": "",
//	  "relation": "",
//	  "subject_type": ""
//	  "subject_id": "",
//	  "trace": false
//	})
func RegisterChecks(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args dsr3.ChecksRequest

			if err := ast.As(op1.Value, &args); err != nil {
				return nil, err
			}

			if proto.Equal(&args, &dsr3.ChecksRequest{}) {
				return helpMsg(fnName, &dsr3.ChecksRequest{
					Default: &dsr3.CheckRequest{
						ObjectType:  "",
						ObjectId:    "",
						Relation:    "",
						SubjectType: "",
						SubjectId:   "",
					},
					Checks: []*dsr3.CheckRequest{
						{
							ObjectType:  "",
							ObjectId:    "",
							Relation:    "",
							SubjectType: "",
							SubjectId:   "",
						},
					},
				})
			}

			if args.Default == nil {
				args.Default = &dsr3.CheckRequest{}
			}

			if args.Checks == nil {
				args.Checks = []*dsr3.CheckRequest{}
			}

			resp, err := dr.GetDS().Checks(bctx.Context, &args)
			if err != nil {
				traceError(&bctx, fnName, err)
				return nil, err
			}

			buf := new(bytes.Buffer)
			if err := ProtoToBuf(buf, resp); err != nil {
				return nil, err
			}

			pbs := structpb.Struct{}
			if err := protojson.Unmarshal(buf.Bytes(), &pbs); err != nil {
				return nil, err
			}

			result := pbs.Fields["checks"].AsInterface().([]interface{})

			v, err := ast.InterfaceToValue(result)
			if err != nil {
				return nil, err
			}

			return ast.NewTerm(v), nil
		}
}
