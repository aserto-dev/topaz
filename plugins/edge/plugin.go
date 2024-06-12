package edge

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/aserto-dev/go-aserto/client"
	"github.com/aserto-dev/go-edge-ds/pkg/datasync"
	"github.com/aserto-dev/go-edge-ds/pkg/directory"
	"github.com/aserto-dev/go-grpc/aserto/api/v2"
	"github.com/aserto-dev/topaz/pkg/app"
	topaz "github.com/aserto-dev/topaz/pkg/cc/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/google/uuid"
	"github.com/open-policy-agent/opa/plugins"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const (
	PluginName    string = "aserto_edge"
	syncScheduler string = "scheduler"
	syncOnDemand  string = "on-demand"
	syncTask      string = "sync-task"
	status        string = "status"
	started       string = "started"
	finished      string = "finished"
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
	ctx            context.Context
	cancel         context.CancelFunc
	manager        *plugins.Manager
	logger         *zerolog.Logger
	config         *Config
	topazConfig    *topaz.Config
	syncNow        chan api.SyncMode
	firstRunSignal chan struct{}
	once           sync.Once
	app            *app.Topaz
}

func newEdgePlugin(logger *zerolog.Logger, cfg *Config, topazConfig *topaz.Config, manager *plugins.Manager, app *app.Topaz) *Plugin {
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
		once:        sync.Once{},
		app:         app,
	}
}

func (p *Plugin) resetContext() {
	p.ctx, p.cancel = context.WithCancel(context.Background())
}

func (p *Plugin) Start(ctx context.Context) error {
	p.logger.Info().Str("id", p.manager.ID).Bool("enabled", p.config.Enabled).Int("interval", p.config.SyncInterval).Msg("EdgePlugin.Start")
	p.manager.UpdatePluginStatus(PluginName, &plugins.Status{State: plugins.StateOK})

	if p.hasLoopBack() {
		p.logger.Warn().
			Str("edge-directory", p.config.Addr).
			Str("remote-directory", p.topazConfig.DirectoryResolver.Address).
			Bool("has-loopback", p.hasLoopBack()).
			Msg("EdgePlugin.Start")
		return nil
	}

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
	if p.config.Enabled != newConfig.Enabled && !p.hasLoopBack() {
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

// A loopback configuration exists when Topaz is configured with a remote directory AND
// an edge sync that points to the same directory instance and tenant as the edge-sync configuration.
// The edge sync can be either explicitly configured in the Topaz configuration file or
// implicitly contributed as part of the discovery response.
// When a loopback is detected, the remote directory configuration takes precedence,
// and the edge sync will be disabled.
func (p *Plugin) hasLoopBack() bool {
	return (p.config.Addr == p.topazConfig.DirectoryResolver.Address &&
		p.config.TenantID == p.topazConfig.DirectoryResolver.TenantID)
}

func (p *Plugin) SyncNow(mode api.SyncMode) {
	p.syncNow <- mode
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

			p.task(api.SyncMode_SYNC_MODE_WATERMARK) // watermark sync

			if cycle%cycles == 0 {
				p.task(api.SyncMode_SYNC_MODE_DIFF)
				cycle = 0
			}
			cycle++

		case mode := <-p.syncNow:
			p.logger.Warn().Time("dispatch", time.Now()).Msg(syncOnDemand)
			interval.Stop()

			p.task(mode)
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

func (p *Plugin) task(mode api.SyncMode) {
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

	conn, err := p.remoteDirectoryClient(ctx)
	if err != nil {
		p.logger.Error().Err(err).Msg(syncTask)
		return
	}
	defer conn.Close()

	opts := []datasync.Option{}
	switch mode {
	case api.SyncMode_SYNC_MODE_UNKNOWN:
		opts = append(opts, datasync.WithMode(datasync.Manifest), datasync.WithMode(datasync.Full))
	case api.SyncMode_SYNC_MODE_FULL:
		opts = append(opts, datasync.WithMode(datasync.Manifest), datasync.WithMode(datasync.Full))
	case api.SyncMode_SYNC_MODE_DIFF:
		opts = append(opts, datasync.WithMode(datasync.Manifest), datasync.WithMode(datasync.Diff))
	case api.SyncMode_SYNC_MODE_WATERMARK:
		opts = append(opts, datasync.WithMode(datasync.Manifest), datasync.WithMode(datasync.Watermark))
	case api.SyncMode_SYNC_MODE_MANIFEST:
		opts = append(opts, datasync.WithMode(datasync.Manifest))
	default:
	}

	ds, err := directory.Get()
	if err != nil {
		p.logger.Error().Err(err).Msg(syncTask)
		return
	}
	if err := ds.DataSyncClient().Sync(ctx, conn, opts...); err != nil {
		p.logger.Error().Err(err).Msg(syncTask)
	}
	if p.config.Enabled {
		p.once.Do(func() {
			p.app.SetServiceStatus("sync", grpc_health_v1.HealthCheckResponse_SERVING)
		})
	}
	p.logger.Info().Str(status, finished).Msg(syncTask)
}

func (p *Plugin) remoteDirectoryClient(ctx context.Context) (*grpc.ClientConn, error) {

	opts := []client.ConnectionOption{
		client.WithAddr(p.config.Addr),
		client.WithInsecure(p.config.Insecure),
	}

	if p.config.APIKey != "" {
		opts = append(opts, client.WithAPIKeyAuth(p.config.APIKey))
	}

	if p.config.TenantID != "" {
		opts = append(opts, client.WithTenantID(p.config.TenantID))
	}

	conn, err := client.NewConnection(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (p *Plugin) GetFirstRunChan() chan struct{} {
	return p.firstRunSignal
}
