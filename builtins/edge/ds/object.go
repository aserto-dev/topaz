package ds

import (
	"bytes"

	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/builtins"
	"github.com/aserto-dev/topaz/resolvers"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/types"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// RegisterObject - ds.object
//
//	ds.object({
//		"object_type": "",
//		"object_id": "",
//		"with_relation": false
//	})
func RegisterObject(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var (
				args struct {
					ObjectType    string `json:"object_type,omitempty"` // v3 object_type
					ObjectID      string `json:"object_id,omitempty"`   // v3 object_id
					WithRelations bool   `json:"with_relations"`        // v3 with_relations (false in case of v2)
				}
				req *reader.GetObjectRequest
			)

			if err := ast.As(op1.Value, &args); err != nil {
				return nil, errors.Wrapf(err, "failed to parse ds.object input message")
			}

			req = &reader.GetObjectRequest{
				ObjectType:    args.ObjectType,
				ObjectId:      args.ObjectID,
				WithRelations: args.WithRelations,
			}

			if proto.Equal(req, &reader.GetObjectRequest{}) {
				return builtins.HelpMsg(fnName, getObjectReq())
			}

			resp, err := dr.GetDS().GetObject(bctx.Context, req)

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
			if err := builtins.ProtoToBuf(buf, resp); err != nil {
				return nil, err
			}

			pbs := structpb.Struct{}
			if err := protojson.Unmarshal(buf.Bytes(), &pbs); err != nil {
				return nil, err
			}

			result, ok := pbs.GetFields()["result"].AsInterface().(map[string]any)
			if !ok {
				return nil, status.Errorf(codes.Internal, "failed type assertion %q", "result")
			}

			relations := pbs.GetFields()["relations"].AsInterface()

			result["relations"] = relations

			v, err := ast.InterfaceToValue(result)
			if err != nil {
				return nil, err
			}

			return ast.NewTerm(v), nil
		}
}

func getObjectReq() *reader.GetObjectRequest {
	return &reader.GetObjectRequest{
		ObjectType:    "",
		ObjectId:      "",
		WithRelations: false,
	}
}
