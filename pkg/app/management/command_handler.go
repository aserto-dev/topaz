package management

import (
	"context"

	"github.com/aserto-dev/go-grpc/aserto/api/v2"
	"github.com/aserto-dev/runtime"
	"github.com/aserto-dev/topaz/plugins/edge"
	"github.com/open-policy-agent/opa/v1/plugins/discovery"
	"github.com/pkg/errors"
)

func HandleCommand(ctx context.Context, cmd *api.Command, r *runtime.Runtime) error {
	switch msg := cmd.GetData().(type) {
	case *api.Command_Discovery:
		plugin := r.GetPluginsManager().Plugin(discovery.Name)
		if plugin == nil {
			return errors.Errorf("failed to find discovery plugin")
		}

		discoveryPlugin, ok := plugin.(*discovery.Discovery)
		if !ok {
			return errors.Errorf("failed to cast discovery plugin")
		}

		err := discoveryPlugin.Trigger(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to trigger discovery")
		}

	case *api.Command_SyncEdgeDirectory:
		plugin := r.GetPluginsManager().Plugin(edge.PluginName)
		if plugin == nil {
			return errors.Errorf("failed to find edge plugin")
		}

		edgePlugin, ok := plugin.(*edge.Plugin)
		if !ok {
			return errors.Errorf("failed to cast edge directory plugin")
		}

		edgePlugin.SyncNow(msg.SyncEdgeDirectory.GetMode())

	default:
		return errors.New("not implemented")
	}

	return nil
}
