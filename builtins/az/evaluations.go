package az

import (
	"github.com/authzen/access.go/api/access/v1"

	"github.com/aserto-dev/topaz/builtins"
	"github.com/aserto-dev/topaz/resolvers"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/types"

	"github.com/rs/zerolog"
	"google.golang.org/protobuf/proto"
)

const azEvaluationsHelp string = `az.evaluations({
	"subject": {"type": "", "id": "", "properties": {}},
	"action": {"name": "", "properties": {}},
	"resource": {"type": "", "id": "", "properties": {}},
  "context": {},
	"options": {},
	"evaluations": [
		{
			"subject": {"type": "", "id": "", "properties": {}},
			"action": {"name": "", "properties": {}},
			"resource": {"type": "", "id": "", "properties": {}},
			"context": {}
		},
		{
			"subject": {"type": "", "id": "", "properties": {}},
			"action": {"name": "", "properties": {}},
			"resource": {"type": "", "id": "", "properties": {}},
			"context": {},
		}
	]
})`

// RegisterEvaluations
// https://openid.github.io/authzen/#name-access-evaluations-api.
func RegisterEvaluations(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args access.EvaluationsRequest

			if err := ast.As(op1.Value, &args); err != nil {
				return nil, err
			}

			if proto.Equal(&args, &access.EvaluationsRequest{}) {
				return ast.StringTerm(azEvaluationsHelp), nil
			}

			resp, err := dr.GetAuthZen().Evaluations(bctx.Context, &args)
			if err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			return builtins.ResponseToTerm(resp)
		}
}
