package runtime

import "github.com/open-policy-agent/opa/v1/ast"

type RegoVersion int

const DefaultRegoVersion = RegoV1

const (
	RegoUndefined RegoVersion = iota
	// RegoV0 is the default, original Rego syntax.
	RegoV0
	// RegoV0CompatV1 requires modules to comply with both the RegoV0 and RegoV1 syntax (as when 'rego.v1' is imported in a module).
	// Shortly, RegoV1 compatibility is required, but 'rego.v1' or 'future.keywords' must also be imported.
	RegoV0CompatV1
	// RegoV1 is the Rego syntax enforced by OPA 1.0; e.g.:
	// future.keywords part of default keyword set, and don't require imports;
	// 'if' and 'contains' required in rule heads;
	// (some) strict checks on by default.
	RegoV1
)

const (
	regoUndefined string = "undefined"
	regoV0        string = "rego.v0"
	regoV0V1      string = "rego.v0v1"
	regoV1        string = "rego.v1"
)

func (v RegoVersion) ToAstRegoVersion() ast.RegoVersion {
	switch v {
	case RegoUndefined:
		return ast.RegoUndefined
	case RegoV0:
		return ast.RegoV0
	case RegoV0CompatV1:
		return ast.RegoV0CompatV1
	case RegoV1:
		return ast.RegoV1
	default:
		return ast.RegoUndefined
	}
}

func (v RegoVersion) String() string {
	switch v {
	case RegoUndefined:
		return regoUndefined
	case RegoV0:
		return regoV0
	case RegoV0CompatV1:
		return regoV0V1
	case RegoV1:
		return regoV1
	default:
		return regoUndefined
	}
}

func RegoVersionFromString(v string) RegoVersion {
	switch v {
	case regoV0:
		return RegoV0
	case regoV0V1:
		return RegoV0CompatV1
	case regoV1:
		return RegoV1
	default:
		return RegoV1
	}
}
