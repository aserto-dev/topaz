package res

import (
	"bytes"
	"encoding/json"

	"github.com/aserto-dev/go-eds/pkg/pb"
	"github.com/aserto-dev/topaz/pkg/app/instance"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
)

// RegisterResList - list resources res.list() []string.
func RegisterResList(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.BuiltinDyn) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction([]types.Type{}, types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, _ []*ast.Term) (*ast.Term, error) {
			instanceID := instance.ExtractID(bctx.Context)

			ds, err := dr.GetDirectory(bctx.Context, instanceID)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get directory")
			}

			keyList, _, _, err := ds.ListResources(instanceID, "", -1) // HACK
			if err != nil {
				return nil, errors.Wrapf(err, "list resources")
			}

			result := struct {
				Keys []string `json:"keys"`
			}{
				Keys: keyList,
			}

			b, err := json.Marshal(result)
			if err != nil {
				return nil, err
			}

			buf := bytes.NewReader(b)

			v, err := ast.ValueFromReader(buf)
			if err != nil {
				return nil, err
			}

			return ast.NewTerm(v), nil
		}
}

// RegisterResGet - get resource by key.
func RegisterResGet(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.S), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, a *ast.Term) (*ast.Term, error) {
			instanceID := instance.ExtractID(bctx.Context)

			var key string
			if err := ast.As(a.Value, &key); err != nil {
				return nil, err
			}

			ds, err := dr.GetDirectory(bctx.Context, instanceID)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get directory")
			}

			value, err := ds.GetResource(instanceID, key)
			if err != nil {
				return nil, errors.Wrapf(err, "get resource [%s]", key)
			}

			buf := new(bytes.Buffer)
			if err := pb.ProtoToBuf(buf, value); err != nil {
				return nil, err
			}

			v, err := ast.ValueFromReader(buf)
			if err != nil {
				return nil, err
			}

			return ast.NewTerm(v), nil
		}
}
