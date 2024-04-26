package datasync

import (
	"context"
	"strings"
	"time"

	"github.com/aserto-dev/go-aserto/client"
	dsc "github.com/aserto-dev/go-directory/pkg/datasync"
	"github.com/aserto-dev/go-edge-ds/pkg/directory"
	topaz "github.com/aserto-dev/topaz/pkg/cc/config"

	"github.com/google/uuid"
	"github.com/open-policy-agent/opa/plugins"
	"github.com/rs/zerolog"
)

const PluginName = "aserto_edge"

const (
	syncScheduler  string = "scheduler"
	syncOnDemand   string = "on-demand"
	syncTask       string = "sync-task"
	syncRun        string = "sync-run"
	syncProducer   string = "producer"
	syncSubscriber string = "subscriber"
	status         string = "status"
	started        string = "started"
	stage          string = "stage"
	finished       string = "finished"
	channelSize    int    = 10000
	localHost      string = "localhost:9292"
)

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

func newDataSyncPlugin(ctx context.Context, logger *zerolog.Logger, cfg *Config, topazConfig *topaz.Config, manager *plugins.Manager) *Plugin {
	newLogger := logger.With().Str("component", "datasync.plugin").Logger()

	cfg.SessionID = uuid.NewString()

	// sync context, lifetime management for scheduler.
	syncContext, cancel := context.WithCancel(ctx)

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

			opts := []dsc.Option{
				dsc.WithMode(dsc.Manifest),
				dsc.WithMode(dsc.Watermark),
			}

			if cycle%cycles == 0 {
				opts = append(opts, dsc.WithMode(dsc.FullDiff))
				cycle = 0
			}
			cycle++

			p.task(opts...)

		case <-p.syncNow:
			p.logger.Warn().Time("dispatch", time.Now()).Msg(syncOnDemand)
			interval.Stop()

			p.task([]dsc.Option{
				dsc.WithMode(dsc.Manifest),
				dsc.WithMode(dsc.Watermark),
				dsc.WithMode(dsc.FullDiff)}...,
			)
		}

		// calculate the interval in secs
		//
		// p.config.SyncInterval 1m-60m
		// 1m -> 60s -> 15s interval
		waitInSec := (p.config.SyncInterval * 60) / cycles

		wait := time.Duration(waitInSec) * time.Second
		interval.Reset(wait)
		p.logger.Info().Str("interval", wait.String()).Time("next-run", time.Now().Add(wait)).Msg(syncScheduler)
	}
}

func (p *Plugin) task(opts ...dsc.Option) error {
	p.logger.Info().Str(status, started).Msg(syncTask)

	defer func() {
		if r := recover(); r != nil {
			p.logger.Error().Interface("recover", r).Msg(syncTask)
		}
	}()

	conn, err := client.NewConnection(p.ctx,
		[]client.ConnectionOption{
			client.WithAddr(p.config.Addr),
			client.WithTenantID(p.config.TenantID),
			client.WithAPIKeyAuth(p.config.APIKey),
			client.WithInsecure(p.config.Insecure),
		}...)
	if err != nil {
		return err
	}
	defer conn.Close()

	d, err := directory.Get()
	if err != nil {
		return err
	}

	if err := d.DataSyncClient().Sync(p.ctx, conn, opts...); err != nil {
		return err
	}

	p.logger.Info().Str(status, finished).Msg(syncTask)

	return nil
}
