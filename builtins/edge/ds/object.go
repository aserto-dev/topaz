package ds

import (
	"bytes"

	dsc2 "github.com/aserto-dev/go-directory/aserto/directory/common/v2"
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
				args     dsr3.GetObjectRequest
				outputV2 bool
			)

			if err := ast.As(op1.Value, &args); err != nil {

				// if v3 input parsing fails, fallback to v2 before exiting with an error.
				var a2 dsc2.ObjectIdentifier
				if err := ast.As(op1.Value, &a2); err != nil {
					return nil, err
				}

				outputV2 = true

				args = dsr3.GetObjectRequest{
					ObjectType:    a2.GetType(),
					ObjectId:      a2.GetKey(),
					WithRelations: false,
				}
			}

			if proto.Equal(&args, &dsr3.GetObjectRequest{}) {
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

			resp, err := client.GetObject(bctx.Context, &args)
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
