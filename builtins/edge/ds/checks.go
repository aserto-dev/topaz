package ds

import (
	"bytes"

	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/builtins"
	"github.com/aserto-dev/topaz/resolvers"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/types"

	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
			var args reader.ChecksRequest

			if err := ast.As(op1.Value, &args); err != nil {
				return nil, err
			}

			if proto.Equal(&args, &reader.ChecksRequest{}) {
				return builtins.HelpMsg(fnName, checksReq())
			}

			if args.GetDefault() == nil {
				args.Default = &reader.CheckRequest{}
			}

			if args.GetChecks() == nil {
				args.Checks = []*reader.CheckRequest{}
			}

			resp, err := dr.GetDS().Checks(bctx.Context, &args)
			if err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			buf := new(bytes.Buffer)
			if err := builtins.ProtoToBuf(buf, resp); err != nil {
				return nil, err
			}

			pbs := structpb.Struct{}
			if err := protojson.Unmarshal(buf.Bytes(), &pbs); err != nil {
				return nil, err
			}

			results, ok := pbs.GetFields()["checks"].AsInterface().([]any)
			if !ok {
				return nil, status.Errorf(codes.Internal, "failed type assertion %q", "results")
			}

			v, err := ast.InterfaceToValue(results)
			if err != nil {
				return nil, err
			}

			return ast.NewTerm(v), nil
		}
}

func checksReq() *reader.ChecksRequest {
	return &reader.ChecksRequest{
		Default: &reader.CheckRequest{
			ObjectType:  "",
			ObjectId:    "",
			Relation:    "",
			SubjectType: "",
			SubjectId:   "",
		},
		Checks: []*reader.CheckRequest{
			{
				ObjectType:  "",
				ObjectId:    "",
				Relation:    "",
				SubjectType: "",
				SubjectId:   "",
			},
		},
	}
}
