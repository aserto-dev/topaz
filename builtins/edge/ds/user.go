package ds

import (
	"bytes"
	"fmt"

	dsc "github.com/aserto-dev/go-directory/aserto/directory/common/v2"
	dsr "github.com/aserto-dev/go-directory/aserto/directory/reader/v2"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/topdown"
	"github.com/open-policy-agent/opa/types"
)

// RegisterUser - ds.user
//
//	ds.user({
//		"id": ""
//	})
func RegisterUser(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {

			type args struct {
				ID string `json:"id"`
			}

			var a args
			if err := ast.As(op1.Value, &a); err != nil {
				return nil, err
			}

			if (args{}) == a {
				return help(fnName, args{})
			}

			client, err := dr.GetDS(bctx.Context)
			if err != nil {
				return nil, errors.Wrapf(err, "get directory client")
			}

			resp, err := client.GetObject(bctx.Context, &dsr.GetObjectRequest{
				Param: &dsc.ObjectIdentifier{
					Id: &a.ID,
				},
			})
			if err != nil {
				if bctx.TraceEnabled {
					if len(bctx.QueryTracers) > 0 {
						bctx.QueryTracers[0].TraceEvent(topdown.Event{
							Op:      topdown.FailOp,
							Message: fmt.Sprintf("DS User Error:%s", err.Error()),
						})
					}
				}
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
