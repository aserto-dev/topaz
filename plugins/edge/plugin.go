package edge

import (
	"context"
	"strings"
	"time"

	topaz "github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/google/uuid"
	"github.com/open-policy-agent/opa/plugins"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const PluginName = "aserto_edge"

type Config struct {
	Enabled      bool   `json:"enabled"`
	Addr         string `json:"addr"`
	APIKey       string `json:"apikey"`
	TenantID     string `json:"tenant_id,omitempty"`
	Timeout      int    `json:"timeout"`
	PageSize     int    `json:"page_size"`
	SyncInterval int    `json:"sync_interval"`
	Insecure     bool   `json:"insecure"`
	SessionID    string `json:"session_id,omitempty"`
}

type Plugin struct {
	ctx         context.Context
	manager     *plugins.Manager
	logger      *zerolog.Logger
	config      *Config
	topazConfig *topaz.Config
	cancel      context.CancelFunc
	syncNow     chan bool
}

func newEdgePlugin(logger *zerolog.Logger, cfg *Config, topazConfig *topaz.Config, manager *plugins.Manager) *Plugin {
	newLogger := logger.With().Str("component", "edge.plugin").Logger()

	cfg.SessionID = uuid.NewString()

	syncContext, cancel := context.WithCancel(context.Background())

	if topazConfig == nil {
		logger.Error().Msg("no topaz directory config was provided")
	}

	return &Plugin{
		ctx:         syncContext,
		logger:      &newLogger,
		manager:     manager,
		config:      cfg,
		topazConfig: topazConfig,
		cancel:      cancel,
		syncNow:     make(chan bool),
	}
}

func (p *Plugin) Start(ctx context.Context) error {
	p.logger.Info().Str("id", p.manager.ID).Bool("enabled", p.config.Enabled).Int("interval", p.config.SyncInterval).Msg("EdgePlugin.Start")

	p.manager.UpdatePluginStatus(PluginName, &plugins.Status{State: plugins.StateOK})

	// start first run after 15 sec delay.
	go p.scheduler(time.NewTicker(
		time.Duration(15) * time.Second),
	)

	return nil
}

func (p *Plugin) Stop(ctx context.Context) {
	p.logger.Info().Str("id", p.manager.ID).Bool("enabled", p.config.Enabled).Int("interval", p.config.SyncInterval).Msg("EdgePlugin.Stop")

	p.cancel()
	p.manager.UpdatePluginStatus(PluginName, &plugins.Status{State: plugins.StateNotReady})
}

func (p *Plugin) Reconfigure(ctx context.Context, config interface{}) {
	p.logger.Trace().Str("id", p.manager.ID).Interface("cur", p.config).Interface("new", config).Msg("EdgePlugin.Reconfigure")

	newConfig := config.(*Config)

	// handle enabled status changed
	if p.config.Enabled != newConfig.Enabled {
		p.logger.Info().Str("id", p.manager.ID).Bool("old", p.config.Enabled).Bool("new", newConfig.Enabled).Msg("sync enabled changed")
		if newConfig.Enabled {
			go p.scheduler(time.NewTicker(
				time.Duration(newConfig.SyncInterval) * time.Minute),
			)
		} else {
			p.cancel()
		}
	}

	p.config = config.(*Config)
	p.config.TenantID = strings.Split(p.manager.ID, "/")[0]
	p.config.SessionID = uuid.NewString()
}

func (p *Plugin) SyncNow() {
	p.syncNow <- true
}

func (p *Plugin) scheduler(interval *time.Ticker) {
	defer interval.Stop()
	wait := time.Duration(p.config.SyncInterval) * time.Minute

	for {
		select {
		case <-p.ctx.Done():
			p.logger.Warn().Time("done", time.Now()).Msg(syncScheduler)
			return
		case t := <-interval.C:
			p.logger.Info().Time("dispatch", t).Msg(syncScheduler)
			interval.Stop()
			p.task()
			interval.Reset(time.Duration(p.config.SyncInterval) * time.Minute)
			p.logger.Info().Str("interval", wait.String()).Time("next-run", time.Now().Add(wait)).Msg(syncScheduler)
		case <-p.syncNow:
			p.logger.Info().Msg("run-now")
			interval.Stop()
			p.task()
			interval.Reset(time.Duration(p.config.SyncInterval) * time.Minute)
			p.logger.Info().Str("interval", wait.String()).Time("next-run", time.Now().Add(wait)).Msg(syncScheduler)
		}
	}
}

func (p *Plugin) task() {
	p.logger.Info().Time("started", time.Now()).Msg("scheduler")

	defer func() {
		if r := recover(); r != nil {
			p.logger.Error().Interface("recover", r).Msg("task-panic")
		}
	}()

	if p.config.TenantID == "" {
		panic(errors.Errorf("tenant-id empty"))
	}

	sync := NewSyncMgr(p.config, p.topazConfig, p.logger)
	sync.Run()

	p.logger.Info().Time("finished", time.Now()).Msg("scheduler")
}
