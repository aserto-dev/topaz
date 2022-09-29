package dir

import (
	"bytes"
	"encoding/json"

	"github.com/aserto-dev/topaz/directory"
	"github.com/aserto-dev/topaz/pkg/app/instance"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
)

// RegisterManagementChain - return management chain as ordered array users for identity.
func RegisterManagementChain(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.S), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, a *ast.Term) (*ast.Term, error) {
			var (
				instanceID = instance.ExtractID(bctx.Context)
				identity   string
			)

			if err := ast.As(a.Value, &identity); err != nil {
				return nil, err
			}

			ds, err := dr.GetDirectory(bctx.Context, instanceID)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get directory")
			}

			uidChain := managementChain(logger, ds, instanceID, identity)

			result := struct {
				Chain []string `json:"chain"`
			}{
				Chain: uidChain,
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

// managementChain - return the transitive management chain for a given identity
// represented as ordered array of user identifiers.
func managementChain(logger *zerolog.Logger, ds directory.Directory, tenantID, ident string) []string {
	logger.Trace().Str("tenantID", tenantID).Str("ident", ident).Msg("ManagementChain")

	results := []string{}

	rootUser, err := ds.GetUserFromIdentity(tenantID, ident)
	if err != nil {
		return []string{}
	}

	current := rootUser

	results = append(results, rootUser.Id)

	for {
		mgrAttr, ok := current.Attributes.Properties.Fields[manager]
		if !ok {
			return []string{}
		}

		mgrUID := mgrAttr.GetStringValue()
		if mgrUID == "" {
			break
		}

		mgrUser, err := ds.GetUserFromIdentity(tenantID, mgrUID)
		if err != nil {
			return []string{}
		}

		results = append(results, mgrUser.Id)

		current = mgrUser
	}

	return results
}
