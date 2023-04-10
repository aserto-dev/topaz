package ds

import (
	"bytes"

	dsc "github.com/aserto-dev/go-directory/aserto/directory/common/v2"
	dsr "github.com/aserto-dev/go-directory/aserto/directory/reader/v2"
	"github.com/aserto-dev/topaz/resolvers"
	"google.golang.org/protobuf/proto"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// RegisterObject - ds.object
//
//	ds.object({
//		"key": "",
//		"type": ""
//	})
func RegisterObject(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: false,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var a *dsc.ObjectIdentifier

			if err := ast.As(op1.Value, &a); err != nil {
				return nil, err
			}

			if a.Type == nil && a.Key == nil {
				a = &dsc.ObjectIdentifier{
					Type: proto.String(""),
					Key:  proto.String(""),
				}
				return help(fnName, a)
			}

			client, err := dr.GetDS(bctx.Context)
			if err != nil {
				return nil, errors.Wrapf(err, "get directory client")
			}

			resp, err := client.GetObject(bctx.Context, &dsr.GetObjectRequest{
				Param: a,
			})
			if err != nil {
				traceError(&bctx, fnName, err)
				return nil, err
			}

			buf := new(bytes.Buffer)
			if resp.Result != nil {
				if err := ProtoToBuf(buf, resp.Result); err != nil {
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
