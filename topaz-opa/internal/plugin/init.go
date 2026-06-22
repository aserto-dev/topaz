package plugin

import (
	"github.com/aserto-dev/topaz/topaz-opa/internal/builtins/ac"
	"github.com/aserto-dev/topaz/topaz-opa/internal/builtins/ds"
	"github.com/open-policy-agent/opa/v1/runtime"
)

const PluginName string = `topaz`

func Init() {
	ac.RegisterAccessBuiltins(GetAccessClient())
	ds.RegisterDirectoryBuiltins(GetDirectoryClient())

	runtime.RegisterPlugin(PluginName, &PluginFactory{})
}
