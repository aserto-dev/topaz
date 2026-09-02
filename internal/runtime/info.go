package runtime

import (
	"os"
	"strings"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/version"
)

func (r *Runtime) info() ast.Object {
	info := ast.NewObject()

	if r.Config != nil {
		v, err := ast.InterfaceToValue(r.Config.Config)
		if err != nil {
			r.Logger.Error().Err(err).Msg("failed to convert config as an opa term")
			return nil
		}

		info.Insert(ast.StringTerm("config"), ast.NewTerm(v))
	}

	env := ast.NewObject()

	r.Logger.Debug().Msg("loading process environment variables as rego terms")

	const maxParts int = 2

	for _, s := range os.Environ() {
		parts := strings.SplitN(s, "=", maxParts)
		if len(parts) == 1 {
			env.Insert(ast.StringTerm(parts[0]), ast.NullTerm())
		} else if len(parts) > 1 {
			env.Insert(ast.StringTerm(parts[0]), ast.StringTerm(parts[1]))
		}
	}

	info.Insert(ast.StringTerm("env"), ast.NewTerm(env))
	info.Insert(ast.StringTerm("version"), ast.StringTerm(version.Version))
	info.Insert(ast.StringTerm("commit"), ast.StringTerm(version.Vcs))

	return info
}
