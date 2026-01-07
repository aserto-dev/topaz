package az

import (
	"github.com/authzen/access.go/api/access/v1"

	"github.com/aserto-dev/topaz/topazd/authorizer/builtins"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/types"

	"github.com/rs/zerolog"
	"google.golang.org/protobuf/proto"
)

const azResourceSearchHelp string = `az.resource_search({
	"subject": {"type": "", "id": "", "properties": {}},
	"action": {"name": "", "properties": {}},
	"resource": {"type": "", "id": "", "properties": {}},
	"context": {},
	"page": {"next_token": ""}
})`

// RegisterResourceSearch.
// https://openid.github.io/authzen/#name-resource-search-api
func RegisterResourceSearch(logger *zerolog.Logger, fnName string, ac access.AccessClient) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args access.ResourceSearchRequest

			if err := ast.As(op1.Value, &args); err != nil {
				return nil, err
			}

			if proto.Equal(&args, &access.ResourceSearchRequest{}) {
				return ast.StringTerm(azResourceSearchHelp), nil
			}

			resp, err := ac.ResourceSearch(bctx.Context, &args)
			if err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			return builtins.ResponseToTerm(resp)
		}
}
