package runtime

import (
	"github.com/open-policy-agent/opa/v1/plugins"
	"github.com/pkg/errors"
)

// CheckPluginsStatus, check all plugins if status is NOT plugins.StateOK.
func (r *Runtime) CheckPluginsStatus() error {
	if r.opaInstance == nil {
		return errors.Errorf("OPA instance not initialized")
	}

	if r.pluginsManager == nil {
		return errors.Errorf("plugin manager not initialized")
	}

	pluginsStatus := r.pluginsManager.PluginStatus()

	for name, status := range pluginsStatus {
		if status.State != plugins.StateOK {
			return errors.Errorf("plugin:%q status:%q", name, status.String())
		}
	}

	return nil
}
