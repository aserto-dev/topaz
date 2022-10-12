package ds

import (
	"bytes"

	v2 "github.com/aserto-dev/go-directory/aserto/directory/common/v2"
	ds2 "github.com/aserto-dev/go-directory/aserto/directory/v2"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/pkg/errors"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"

	"github.com/rs/zerolog"
)

// RegisterRelation - ds.relation
//
// ds.check_relation({
//     "object_id": "",
//     "object_type": "",
//     "relation": "",
//     "subject_id": "",
//     "subject_type": ""
//   })
//
func RegisterRelation(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: false,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var a *v2.RelationIdentifier
			if err := ast.As(op1.Value, &a); err != nil {
				return nil, err
			}

			if !ValidateRelation(a) {
				return nil, errors.Errorf("invalid relation arguments")
			}

			client, err := dr.GetDS(bctx.Context)
			if err != nil {
				return nil, errors.Wrapf(err, "get directory client")
			}

			resp, err := client.GetRelation(bctx.Context, &ds2.GetRelationRequest{
				Param: a,
			})
			if err != nil {
				return nil, err
			}

			buf := new(bytes.Buffer)
			if resp != nil {
				if err := ProtoToBuf(buf, resp); err != nil {
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
