package plugin

import (
	"github.com/aserto-dev/topaz/topaz-opa/internal/builtins/ac"
	"github.com/aserto-dev/topaz/topaz-opa/internal/builtins/ds"
	"github.com/aserto-dev/topaz/topaz-opa/internal/plugins/topaz"

	"github.com/open-policy-agent/opa/v1/runtime"
)

func Init() {
	ac.RegisterAccessBuiltins(topaz.GetAccessClient())
	ds.RegisterDirectoryBuiltins(topaz.GetDirectoryClient())

	runtime.RegisterPlugin(topaz.PluginName, &topaz.PluginFactory{})
}
