package noop

import (
	"context"

	"github.com/open-policy-agent/opa/plugins"
)

type Noop struct {
	Manager *plugins.Manager
	Name    string
}

func (dl *Noop) Start(ctx context.Context) error {
	dl.Manager.UpdatePluginStatus(dl.Name, &plugins.Status{State: plugins.StateOK})
	return nil
}

func (dl *Noop) Stop(ctx context.Context) {
	dl.Manager.UpdatePluginStatus(dl.Name, &plugins.Status{State: plugins.StateNotReady})
}

func (dl *Noop) Reconfigure(ctx context.Context, config interface{}) {
}
