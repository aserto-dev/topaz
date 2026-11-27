package ds

import (
	"bytes"

	"github.com/aserto-dev/topaz/api/directory/v3/reader"
	"github.com/aserto-dev/topaz/internal/builtins"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// RegisterUser - ds.user.
func RegisterUser(fnName string, dr func() (reader.ReaderClient, error)) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args struct {
				ID string `json:"id"`
			}

			dsr, err := dr()
			if err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			if err := ast.As(op1.Value, &args); err != nil {
				return nil, err
			}

			if args.ID == "" {
				return userTemplate()
			}

			resp, err := dsr.GetObject(bctx.Context, &reader.GetObjectRequest{
				ObjectType:    "user",
				ObjectId:      args.ID,
				WithRelations: false,
			})

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

			var result proto.Message

			if resp.GetResult() != nil {
				result = resp.GetResult()
			}

			if err := builtins.ProtoToBuf(buf, result); err != nil {
				return nil, err
			}

			v, err := ast.ValueFromReader(buf)
			if err != nil {
				return nil, err
			}

			return ast.NewTerm(v), nil
		}
}

func userTemplate() (*ast.Term, error) {
	t := map[string]any{
		"id": "",
	}

	val, err := ast.InterfaceToValue(t)
	if err != nil {
		return nil, err
	}

	return ast.NewTerm(val), nil
}
