package topaz_file_decision_logger

import (
	"context"
	"os"

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
	logger     *zerolog.Logger
	fileLogger *lumberjack.Logger
	dlogger    zerolog.Logger
}

var _ plugins.Plugin = (*Plugin)(nil)

func (p *Plugin) Start(ctx context.Context) error {
	p.logger.Info().Bool("enabled", p.config.Enabled).Msgf("start")

	if os.Getenv("TOPAZ_RUNNING_IN_CONTAINER") == "true" {
		p.logger.Info().Bool("enabled", p.config.Enabled).Str("file", p.config.Logger.Filename).Msg("running in container")
	}

	p.fileLogger = &lumberjack.Logger{
		Filename:   p.config.Logger.Filename,
		MaxSize:    p.config.Logger.MaxSize,
		MaxBackups: p.config.Logger.MaxBackups,
		MaxAge:     p.config.Logger.MaxAge,
		LocalTime:  p.config.Logger.LocalTime,
		Compress:   p.config.Logger.Compress,
	}

	// verify ability to write from the fileLogger, before handing it to zerolog.
	if _, err := p.fileLogger.Write([]byte{}); err != nil {
		p.logger.Error().Bool("enabled", p.config.Enabled).Str("file", p.fileLogger.Filename).Err(err).Msg("file-logger write failed")
		return err
	}

	p.dlogger = zerolog.New(p.fileLogger)

	p.manager.UpdatePluginStatus(PluginName, &plugins.Status{State: plugins.StateOK})

	p.logger.Info().Bool("enabled", p.config.Enabled).Str("file", p.fileLogger.Filename).Msg("started")

	return nil
}

func (p *Plugin) Stop(ctx context.Context) {
	p.logger.Info().Bool("enabled", p.config.Enabled).Str("file", p.fileLogger.Filename).Msg("stop")

	if p.fileLogger != nil {
		_ = p.fileLogger.Close()
		p.fileLogger = nil
		p.dlogger = zerolog.Nop()
	}

	p.manager.UpdatePluginStatus(PluginName, &plugins.Status{State: plugins.StateNotReady})

	p.logger.Info().Bool("enabled", p.config.Enabled).Msg("stopped")
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
