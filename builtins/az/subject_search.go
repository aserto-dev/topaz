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

const azSubjectSearchHelp string = `az.subject_search({
	"subject": {"type": "", "properties": {}},
	"action": {"name": "", "properties": {}},
	"resource": {"type": "", "id": "", "properties": {}},
	"context": {},
	"page": {"next_token": ""}
})`

// RegisterSubjectSearch, note: subject_search omits `subject.id` fields when submitted.
// https://openid.github.io/authzen/#name-subject-search-api.
func RegisterSubjectSearch(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args access.SubjectSearchRequest

			if err := ast.As(op1.Value, &args); err != nil {
				return nil, err
			}

			if proto.Equal(&args, &access.SubjectSearchRequest{}) {
				return ast.StringTerm(azSubjectSearchHelp), nil
			}

			resp, err := dr.GetAuthZen().SubjectSearch(bctx.Context, &args)
			if err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			return builtins.ResponseToTerm(resp)
		}
}
