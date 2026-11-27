package ds

import (
	"github.com/aserto-dev/topaz/api/directory/v3/reader"
	"github.com/aserto-dev/topaz/internal/builtins"

	"github.com/open-policy-agent/opa/v1/rego"
)

// RegisterDirectoryBuiltins register Topaz Directory Service builtins.
func RegisterDirectoryBuiltins(dsClient func() (reader.ReaderClient, error)) {
	rego.RegisterBuiltin1(RegisterObject(builtins.DSObject, dsClient))
	rego.RegisterBuiltin1(RegisterRelation(builtins.DSRelation, dsClient))
	rego.RegisterBuiltin1(RegisterRelations(builtins.DSRelations, dsClient))
	rego.RegisterBuiltin1(RegisterGraph(builtins.DSGraph, dsClient))
	rego.RegisterBuiltin1(RegisterCheck(builtins.DSCheck, dsClient))
	rego.RegisterBuiltin1(RegisterChecks(builtins.DSChecks, dsClient))
}
