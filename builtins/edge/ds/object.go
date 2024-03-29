package ds

import (
	"bytes"

	dsc3 "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/go-directory/pkg/convert"
	"github.com/aserto-dev/topaz/resolvers"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// RegisterObject - ds.object
//
// v3 (latest) request format:
//
//	ds.object({
//		"object_type": "",
//		"object_id": "",
//		"with_relation": false
//	})
//
// v2 request format:
//
//	ds.object({
//		"type": "",
//		"key": ""
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
					Type          string `json:"type,omitempty"`        // v2 type
					Key           string `json:"key,omitempty"`         // v2 key
				}
				outputV2 bool
				req      *dsr3.GetObjectRequest
			)

			if err := ast.As(op1.Value, &args); err != nil {
				return nil, errors.Wrapf(err, "failed to parse ds.object input message")
			}

			req = &dsr3.GetObjectRequest{
				ObjectType:    args.ObjectType,
				ObjectId:      args.ObjectID,
				WithRelations: args.WithRelations,
			}

			if args.ObjectType == "" && args.ObjectID == "" && args.Type != "" && args.Key != "" {
				req = &dsr3.GetObjectRequest{
					ObjectType:    args.Type,
					ObjectId:      args.Key,
					WithRelations: false,
				}
				outputV2 = true
			}

			if proto.Equal(req, &dsr3.GetObjectRequest{}) {
				return helpMsg(fnName, &dsr3.GetObjectRequest{
					ObjectType:    "",
					ObjectId:      "",
					WithRelations: false,
				})
			}

			client, err := dr.GetDS(bctx.Context)
			if err != nil {
				return nil, errors.Wrapf(err, "get directory client")
			}

			resp, err := client.GetObject(bctx.Context, req)
			switch {
			case status.Code(err) == codes.NotFound:
				traceError(&bctx, fnName, err)
				astVal, err := ast.InterfaceToValue(map[string]any{})
				if err != nil {
					return nil, err
				}
				return ast.NewTerm(astVal), nil
			case err != nil:
				return nil, err
			}

			if outputV2 {
				v, err := v2Value(resp.Result)
				if err != nil {
					return nil, err
				}

				return ast.NewTerm(v), nil
			}

			buf := new(bytes.Buffer)
			if err := ProtoToBuf(buf, resp); err != nil {
				return nil, err
			}

			pbs := structpb.Struct{}
			if err := protojson.Unmarshal(buf.Bytes(), &pbs); err != nil {
				return nil, err
			}

			result := pbs.Fields["result"].AsInterface().(map[string]interface{})
			relations := pbs.Fields["relations"].AsInterface()
			result["relations"] = relations

			v, err := ast.InterfaceToValue(result)
			if err != nil {
				return nil, err
			}

			return ast.NewTerm(v), nil
		}
}

func v2Value(obj *dsc3.Object) (ast.Value, error) {
	buf := new(bytes.Buffer)

	v2 := convert.ObjectToV2(obj)
	if err := ProtoToBuf(buf, v2); err != nil {
		return nil, err
	}

	return ast.ValueFromReader(buf)
}
