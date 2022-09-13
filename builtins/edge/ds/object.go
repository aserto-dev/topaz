package ds

import (
	"bytes"

	ds2 "github.com/aserto-dev/go-directory/aserto/directory/v2"
	"github.com/aserto-dev/go-eds/pkg/pb"
	"github.com/aserto-dev/topaz/resolvers"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// RegisterObject - ds.object
//
// ds.object({
//     "id": ""
//   })
//
// ds.object({
//     "key": "",
//     "type": ""
//   })
//
func RegisterObject(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: false,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var a ObjectParam

			if err := ast.As(op1.Value, &a); err != nil {
				return nil, err
			}

			if (ObjectParam{}) == a {
				return help(fnName, ObjectParam{})
			}

			client, err := dr.GetDS(bctx.Context)
			if err != nil {
				return nil, errors.Wrapf(err, "get directory client")
			}

			objectParam := a.Validate()
			if objectParam == nil {
				return nil, errors.Errorf("invalid object arguments")
			}

			resp, err := client.GetObject(bctx.Context, &ds2.GetObjectRequest{
				Param: objectParam,
			})
			if err != nil {
				return nil, err
			}

			buf := new(bytes.Buffer)
			if len(resp.Results) == 1 {
				if err := pb.ProtoToBuf(buf, resp.Results[0]); err != nil {
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
