package ac

import (
	"github.com/aserto-dev/topaz/topaz-opa/internal/builtins"
	"github.com/authzen/access.go/api/access/v1"

	"github.com/open-policy-agent/opa/v1/rego"
)

// RegisterAccessBuiltins register the AuthZen Access Service built-ins.
func RegisterAccessBuiltins(acClient func() (access.AccessClient, error)) {
	rego.RegisterBuiltin1(registerEvaluation(builtins.AZEvaluation, acClient))
	rego.RegisterBuiltin1(registerEvaluations(builtins.AZEvaluations, acClient))
	rego.RegisterBuiltin1(registerSubjectSearch(builtins.AZSubjectSearch, acClient))
	rego.RegisterBuiltin1(registerResourceSearch(builtins.AZResourceSearch, acClient))
	rego.RegisterBuiltin1(registerActionSearch(builtins.AZActionSearch, acClient))
}
