package ds

import (
	"github.com/aserto-dev/topaz/topaz-opa/internal/builtins"

	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/open-policy-agent/opa/v1/rego"
)

// RegisterDirectoryBuiltins register Topaz Directory Service builtins.
func RegisterDirectoryBuiltins(dsClient func() reader.ReaderClient) {
	rego.RegisterBuiltin1(registerObject(builtins.DSObject, dsClient))
	rego.RegisterBuiltin1(registerRelation(builtins.DSRelation, dsClient))
	rego.RegisterBuiltin1(registerRelations(builtins.DSRelations, dsClient))
	rego.RegisterBuiltin1(registerGraph(builtins.DSGraph, dsClient))
	rego.RegisterBuiltin1(registerCheck(builtins.DSCheck, dsClient))
	rego.RegisterBuiltin1(registerChecks(builtins.DSChecks, dsClient))
}
