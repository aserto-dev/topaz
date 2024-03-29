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
	cancel      context.CancelFunc
	manager     *plugins.Manager
	logger      *zerolog.Logger
	config      *Config
	topazConfig *topaz.Config
	syncNow     chan bool
}

func newEdgePlugin(logger *zerolog.Logger, cfg *Config, topazConfig *topaz.Config, manager *plugins.Manager) *Plugin {
	newLogger := logger.With().Str("component", "edge.plugin").Logger()

	cfg.SessionID = uuid.NewString()

	// sync context, lifetime management for scheduler.
	syncContext, cancel := context.WithCancel(context.Background())

	if topazConfig == nil {
		logger.Error().Msg("no topaz directory config was provided")
	}

	return &Plugin{
		ctx:         syncContext,
		cancel:      cancel,
		logger:      &newLogger,
		manager:     manager,
		config:      cfg,
		topazConfig: topazConfig,
		syncNow:     make(chan bool),
	}
}

func (p *Plugin) resetContext() {
	p.ctx, p.cancel = context.WithCancel(context.Background())
}

func (p *Plugin) Start(ctx context.Context) error {
	p.logger.Info().Str("id", p.manager.ID).Bool("enabled", p.config.Enabled).Int("interval", p.config.SyncInterval).Msg("EdgePlugin.Start")

	p.manager.UpdatePluginStatus(PluginName, &plugins.Status{State: plugins.StateOK})

	go p.scheduler()

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
			p.resetContext()
			go p.scheduler()
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

const cycles int = 4

func (p *Plugin) scheduler() {
	// scheduler startup delay 15s
	interval := time.NewTicker(15 * time.Second)
	defer interval.Stop()

	cycle := cycles

	for {
		select {
		case <-p.ctx.Done():
			p.logger.Warn().Time("done", time.Now()).Msg(syncScheduler)
			return

		case t := <-interval.C:
			p.logger.Info().Time("dispatch", t).Msg(syncScheduler)
			interval.Stop()

			p.task(false) // watermark sync

			if cycle%cycles == 0 {
				p.task(cycle%cycles == 0)
				cycle = 0
			}
			cycle++

		case <-p.syncNow:
			p.logger.Warn().Time("dispatch", time.Now()).Msg(syncOnDemand)
			interval.Stop()

			p.task(false) // watermark sync
			p.task(true)  // full-diff sync
		}

		// calculate the interval in secs
		//
		// p.config.SyncInterval 1m-60m
		// 1m -> 60s -> 15s interval
		// 5m -> 300s -> 75s interval
		// 60m -> 3600s -> 900s interval
		waitInSec := (p.config.SyncInterval * 60) / cycles

		wait := time.Duration(waitInSec) * time.Second
		interval.Reset(wait)
		p.logger.Info().Str("interval", wait.String()).Time("next-run", time.Now().Add(wait)).Msg(syncScheduler)
	}
}

func (p *Plugin) task(fullSync bool) {
	p.logger.Info().Str(status, started).Msg(syncTask)

	defer func() {
		if r := recover(); r != nil {
			p.logger.Error().Interface("recover", r).Msg(syncTask)
		}
	}()

	if p.config.TenantID == "" {
		panic(errors.Errorf("tenant-id empty"))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		p.logger.Trace().Msg("task cleanup")
		cancel()
	}()

	sync := NewSyncMgr(ctx, p.config, p.topazConfig, p.logger)
	sync.Run(fullSync)
	sync = nil

	p.logger.Info().Str(status, finished).Msg(syncTask)
}
