package topaz_file_decision_logger

import (
	"context"
	"encoding/json"

	api "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/open-policy-agent/opa/v1/plugins/logs"
	"github.com/pkg/errors"
)

func (plugin *Plugin) LogDecision(ctx context.Context, d *api.Decision) error {
	if !plugin.config.Enabled || plugin.fileLogger == nil {
		return nil
	}

	d.Policy.RegistryService = plugin.config.PolicyInfo.RegistryService
	d.Policy.RegistryImage = plugin.config.PolicyInfo.RegistryImage
	d.Policy.RegistryTag = plugin.config.PolicyInfo.RegistryTag
	d.Policy.RegistryDigest = plugin.config.PolicyInfo.Digest

	bytes, err := json.Marshal(d)
	if err != nil {
		return errors.Wrap(err, "error marshaling decision")
	}

	plugin.dlogger.Log().Msg(string(bytes))

	return nil
}

func (plugin *Plugin) Log(ctx context.Context, event logs.EventV1) error {
	if !plugin.config.Enabled || plugin.fileLogger == nil {
		return nil
	}

	return nil
}
