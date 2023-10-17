package ds

import (
	"bytes"

	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/go-edge-ds/pkg/convert"
	"github.com/aserto-dev/topaz/resolvers"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/proto"
)

// RegisterUser - ds.user
//
// v3 (latest) request format:
//
//	ds.user({
//		"id": ""
//	})
//
// v2 request format:
//
//	ds.user({
//		"key": ""
//	})
func RegisterUser(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {

			type argsV3 struct {
				ID string `json:"id"`
			}

			var (
				args struct {
					ID  string `json:"id"`
					Key string `json:"key"`
				}
				outputV2 bool
			)

			if err := ast.As(op1.Value, &args); err != nil {
				return nil, err
			}

			if args.ID == "" && args.Key != "" {
				outputV2 = true
				args.ID = args.Key
			}

			if args.ID == "" && args.Key == "" {
				return help(fnName, argsV3{})
			}

			client, err := dr.GetDS(bctx.Context)
			if err != nil {
				return nil, errors.Wrapf(err, "get directory client")
			}

			resp, err := client.GetObject(bctx.Context, &dsr3.GetObjectRequest{
				ObjectType:    "user",
				ObjectId:      args.ID,
				WithRelations: false,
			})
			if err != nil {
				traceError(&bctx, fnName, err)
				return nil, err
			}

			buf := new(bytes.Buffer)
			var result proto.Message

			if resp.Result != nil {
				result = resp.Result
				if outputV2 {
					result = convert.ObjectToV2(resp.Result)
				}

				if err := ProtoToBuf(buf, result); err != nil {
					return nil, err
				}
			}

			v, err := ast.ValueFromReader(buf)
			if err != nil {
				return nil, err
			}

			return ast.NewTerm(v), nil
		}
}
