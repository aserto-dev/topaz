package authzen_logger

import (
	"context"

	"github.com/open-policy-agent/opa/v1/plugins"
	"gopkg.in/natefinch/lumberjack.v2"
)

const PluginName = "authzen_logger"

type Plugin struct {
	manager    *plugins.Manager
	config     *Config
	fileLogger *lumberjack.Logger
}

var _ plugins.Plugin = (*Plugin)(nil)

func (p *Plugin) Start(ctx context.Context) error {
	p.manager.Logger().Info("AuthZEN logger plugin started")

	p.fileLogger = &lumberjack.Logger{
		Filename:   p.config.Logger.Filename,
		MaxSize:    p.config.Logger.MaxSize,
		MaxBackups: p.config.Logger.MaxBackups,
		MaxAge:     p.config.Logger.MaxAge,
		LocalTime:  p.config.Logger.LocalTime,
		Compress:   p.config.Logger.Compress,
	}

	p.manager.UpdatePluginStatus(PluginName, &plugins.Status{State: plugins.StateOK})

	return nil
}

func (p *Plugin) Stop(ctx context.Context) {
	p.manager.Logger().Info("AuthZEN logger plugin stopped")

	if p.fileLogger != nil {
		_ = p.fileLogger.Close()
	}
}

func (p *Plugin) Reconfigure(ctx context.Context, c any) {
	if cfg, ok := c.(*Config); ok {
		p.config = cfg
	}
}
