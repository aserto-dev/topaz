package ac

import (
	"github.com/aserto-dev/topaz/internal/builtins"
	"github.com/authzen/access.go/api/access/v1"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/types"
	"google.golang.org/protobuf/proto"
)

// RegisterEvaluations
// https://openid.github.io/authzen/#name-access-evaluations-api
func RegisterEvaluations(fnName string, ac func() (access.AccessClient, error)) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args access.EvaluationsRequest

			acr, err := ac()
			if err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			if err := ast.As(op1.Value, &args); err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			if proto.Equal(&args, &access.EvaluationsRequest{}) {
				return evaluationsTemplate()
			}

			resp, err := acr.Evaluations(bctx.Context, &args)
			if err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			return builtins.ResponseToTerm(resp)
		}
}

//nolint:goconst
func evaluationsTemplate() (*ast.Term, error) {
	t := map[string]any{
		"subject":  map[string]any{"type": "", "id": "", "properties": map[string]any{}},
		"action":   map[string]any{"name": "", "properties": map[string]any{}},
		"resource": map[string]any{"type": "", "id": "", "properties": map[string]any{}},
		"context":  map[string]any{},
		"options":  map[string]any{},
		"evaluations": []map[string]any{
			{
				"subject":  map[string]any{"type": "", "id": "", "properties": map[string]any{}},
				"action":   map[string]any{"name": "", "properties": map[string]any{}},
				"resource": map[string]any{"type": "", "id": "", "properties": map[string]any{}},
				"context":  map[string]any{},
			},
		},
	}

	val, err := ast.InterfaceToValue(t)
	if err != nil {
		return nil, err
	}

	return ast.NewTerm(val), nil
}
