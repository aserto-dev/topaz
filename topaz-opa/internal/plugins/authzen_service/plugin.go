package authzen_service

import (
	"context"

	"github.com/open-policy-agent/opa/v1/plugins"
)

const PluginName string = `authzen_service`

type Plugin struct {
	manager *plugins.Manager
	config  *Config
}

var _ plugins.Plugin = (*Plugin)(nil)

func (p *Plugin) Start(ctx context.Context) error {
	p.manager.Logger().Info("authzen service plugin started")

	return nil
}

func (p *Plugin) Stop(ctx context.Context) {
	p.manager.Logger().Info("authzen service plugin stopped")
}

func (p *Plugin) Reconfigure(ctx context.Context, c any) {
	if cfg, ok := c.(*Config); ok {
		p.config = cfg
	}
}
