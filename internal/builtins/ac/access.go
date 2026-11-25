package ac

import (
	"github.com/aserto-dev/topaz/internal/builtins"
	"github.com/authzen/access.go/api/access/v1"

	"github.com/open-policy-agent/opa/v1/rego"
)

const (
	defPageLimit int = 100
)

// RegisterAccessBuiltins register the AuthZen Access Service built-ins.
func RegisterAccessBuiltins(acClient func() (access.AccessClient, error)) {
	rego.RegisterBuiltin1(RegisterEvaluation(builtins.AZEvaluation, acClient))
	rego.RegisterBuiltin1(RegisterEvaluations(builtins.AZEvaluations, acClient))
	rego.RegisterBuiltin1(RegisterSubjectSearch(builtins.AZSubjectSearch, acClient))
	rego.RegisterBuiltin1(RegisterResourceSearch(builtins.AZResourceSearch, acClient))
	rego.RegisterBuiltin1(RegisterActionSearch(builtins.AZActionSearch, acClient))
}
