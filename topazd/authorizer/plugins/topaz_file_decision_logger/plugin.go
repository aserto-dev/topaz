package topaz_file_decision_logger

import (
	"context"

	"github.com/open-policy-agent/opa/v1/plugins"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	PluginName = "topaz_file_decision_logger"
	PluginDesc = "Topaz File Decision Logger"
)

type Plugin struct {
	manager    *plugins.Manager
	config     *Config
	fileLogger *lumberjack.Logger
	dlogger    zerolog.Logger
}

var _ plugins.Plugin = (*Plugin)(nil)

func (p *Plugin) Start(ctx context.Context) error {
	p.manager.Logger().Info(PluginDesc + " started")

	p.fileLogger = &lumberjack.Logger{
		Filename:   p.config.Logger.Filename,
		MaxSize:    p.config.Logger.MaxSize,
		MaxBackups: p.config.Logger.MaxBackups,
		MaxAge:     p.config.Logger.MaxAge,
		LocalTime:  p.config.Logger.LocalTime,
		Compress:   p.config.Logger.Compress,
	}

	p.dlogger = zerolog.New(p.fileLogger)

	p.manager.UpdatePluginStatus(PluginName, &plugins.Status{State: plugins.StateOK})

	return nil
}

func (p *Plugin) Stop(ctx context.Context) {
	p.manager.Logger().Info(PluginDesc + " stopped")

	if p.fileLogger != nil {
		_ = p.fileLogger.Close()
		p.fileLogger = nil
		p.dlogger = zerolog.Nop()
	}
}

func (p *Plugin) Reconfigure(ctx context.Context, c any) {
	if cfg, ok := c.(*Config); ok {
		p.config = cfg
	}
}

func Lookup(m *plugins.Manager) *Plugin {
	p := m.Plugin(PluginName)
	if p == nil {
		return nil
	}

	plugin, _ := p.(*Plugin)

	return plugin
}
