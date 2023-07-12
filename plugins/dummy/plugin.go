package dummy

import (
	"context"

	"github.com/open-policy-agent/opa/plugins"
)

type Dummy struct {
	Manager *plugins.Manager
	Name    string
}

func (dl *Dummy) Start(ctx context.Context) error {
	dl.Manager.UpdatePluginStatus(dl.Name, &plugins.Status{State: plugins.StateOK})
	return nil
}

func (dl *Dummy) Stop(ctx context.Context) {
	dl.Manager.UpdatePluginStatus(dl.Name, &plugins.Status{State: plugins.StateNotReady})
}

func (dl *Dummy) Reconfigure(ctx context.Context, config interface{}) {

}
