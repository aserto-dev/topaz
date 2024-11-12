package edge

import (
	"context"
	"strings"
	"sync/atomic"
	"time"

	client "github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/go-edge-ds/pkg/datasync"
	"github.com/aserto-dev/go-edge-ds/pkg/directory"
	"github.com/aserto-dev/go-grpc/aserto/api/v2"
	"github.com/aserto-dev/topaz/pkg/app"
	topaz "github.com/aserto-dev/topaz/pkg/cc/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
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
	Enabled           bool          `json:"enabled"`
	Addr              string        `json:"addr"`
	APIKey            string        `json:"apikey"`
	TenantID          string        `json:"tenant_id,omitempty"`
	Timeout           int           `json:"timeout"`
	SyncInterval      int           `json:"sync_interval"`
	Insecure          bool          `json:"insecure"`
	SessionID         string        `json:"session_id,omitempty"`
	ConnectionTimeout time.Duration `json:"-"`
	// Deprecated: No longer used.
	PageSize int `json:"page_size,omitempty"`
}

type Plugin struct {
	ctx         context.Context
	cancel      context.CancelFunc
	manager     *plugins.Manager
	logger      *zerolog.Logger
	config      *Config
	topazConfig *topaz.Config
	syncNow     chan api.SyncMode
}

func newEdgePlugin(logger *zerolog.Logger, cfg *Config, topazConfig *topaz.Config, manager *plugins.Manager) *Plugin {
	newLogger := logger.With().Str("component", "edge.plugin").Logger()

	cfg.SessionID = uuid.NewString()
	cfg.ConnectionTimeout = time.Duration(cfg.Timeout * int(time.Second))

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
		syncNow:     make(chan api.SyncMode),
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
			// set health status to NOT_SERVING when plugin switches to disabled.
			app.SetServiceStatus(p.logger, "sync", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
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
	// scheduler startup delay 1s
	interval := time.NewTicker(1 * time.Second)
	defer interval.Stop()

	var running atomic.Bool
	running.Store(false)

	cycle := cycles

	intervalMode := api.SyncMode_SYNC_MODE_UNKNOWN
	onDemandMode := api.SyncMode_SYNC_MODE_UNKNOWN

	for {
		select {
		case <-p.ctx.Done():
			p.logger.Debug().Time("done", time.Now()).Msg(syncScheduler)
			return

		case t := <-interval.C:
			p.logger.Info().Time("dispatch", t).Msg(syncScheduler)
			interval.Stop()

			intervalMode = api.SyncMode_SYNC_MODE_WATERMARK

			if cycle%cycles == 0 {
				intervalMode = api.SyncMode_SYNC_MODE_DIFF
				cycle = 0
			}
			cycle++
			p.logger.Debug().Str("mode", printMode(intervalMode)).Msg("interval handler")

		case mode := <-p.syncNow:
			p.logger.Info().Time("dispatch", time.Now()).Msg(syncOnDemand)
			interval.Stop()

			onDemandMode = fold(onDemandMode, mode)
			p.logger.Debug().Str("mode", printMode(onDemandMode)).Msg("on-demand handler")
		}

		if !running.Load() {
			// determine the run mode
			runMode := api.SyncMode_SYNC_MODE_UNKNOWN
			if onDemandMode != api.SyncMode_SYNC_MODE_UNKNOWN {
				runMode = onDemandMode
				onDemandMode = api.SyncMode_SYNC_MODE_UNKNOWN
			} else {
				runMode = intervalMode
			}

			go func() {
				p.logger.Debug().Str("mode", printMode(runMode)).Msg("start task")

				running.Store(true)
				defer func() {
					p.logger.Debug().Str("mode", printMode(runMode)).Msg("finished task")
					running.Store(false)
				}()

				p.task(runMode)

				// if on-demand mode is UNKNOWN, meaning no new on-demand requests were received while processing the last run, fall back to interval mode.
				if onDemandMode == api.SyncMode_SYNC_MODE_UNKNOWN {
					wait := p.calcInterval()
					interval.Reset(wait)
					p.logger.Info().Str("interval", wait.String()).Time("next-run", time.Now().Add(wait)).Msg(syncScheduler)
				} else {
					p.logger.Warn().Str("mode", printMode(onDemandMode)).Msg("trigger queued on-demand mode")
					p.SyncNow(onDemandMode)
				}
			}()
		}
	}
}

// calcInterval - calculates the next time interval in secs,
// based on the configuration SyncInterval (defined on the EdgeDirectory connection)
// returning a time.Duration.
//
// p.config.SyncInterval 1m-60m
// 1m -> 60s -> 15s interval
// 60m -> 3600s -> 900s interval.
func (p *Plugin) calcInterval() time.Duration {
	waitInSec := (p.config.SyncInterval * 60) / cycles
	return time.Duration(waitInSec) * time.Second
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

	conn, err := p.remoteDirectoryClient()
	if err != nil {
		p.logger.Error().Err(err).Msg(syncTask)
		return
	}
	defer conn.Close()

	ds, err := directory.Get()
	if err != nil {
		p.logger.Error().Err(err).Msg(syncTask)
		return
	}

	if mode == api.SyncMode_SYNC_MODE_UNKNOWN {
		p.exec(ctx, ds, conn, []datasync.Option{
			datasync.WithMode(datasync.Manifest),
			datasync.WithMode(datasync.Full),
		})
		return
	}

	if has(mode, api.SyncMode_SYNC_MODE_WATERMARK) {
		p.exec(ctx, ds, conn, []datasync.Option{
			datasync.WithMode(datasync.Manifest),
			datasync.WithMode(datasync.Watermark),
		})
	}

	if has(mode, api.SyncMode_SYNC_MODE_DIFF) {
		p.exec(ctx, ds, conn, []datasync.Option{
			datasync.WithMode(datasync.Manifest),
			datasync.WithMode(datasync.Diff),
		})
		return
	}

	if has(mode, api.SyncMode_SYNC_MODE_FULL) {
		p.exec(ctx, ds, conn, []datasync.Option{
			datasync.WithMode(datasync.Manifest),
			datasync.WithMode(datasync.Full),
		})
		return
	}

	if has(mode, api.SyncMode_SYNC_MODE_MANIFEST) && !has(mode, api.SyncMode_SYNC_MODE_WATERMARK) {
		p.exec(ctx, ds, conn, []datasync.Option{
			datasync.WithMode(datasync.Manifest),
		})
		return
	}
}

func (p *Plugin) exec(ctx context.Context, ds *directory.Directory, conn *grpc.ClientConn, opts []datasync.Option) {
	if err := ds.DataSyncClient().Sync(ctx, conn, opts...); err != nil {
		p.logger.Error().Err(err).Msg(syncTask)
	}

	if p.config.Enabled {
		app.SetServiceStatus(p.logger, "sync", grpc_health_v1.HealthCheckResponse_SERVING)
	}

	p.logger.Info().Str(status, finished).Msg(syncTask)
}

func (p *Plugin) remoteDirectoryClient() (*grpc.ClientConn, error) {
	cfg := &client.Config{
		Address:  p.config.Addr,
		Insecure: p.config.Insecure,
		APIKey:   p.config.APIKey,
		TenantID: p.config.TenantID,
		Headers:  p.topazConfig.DirectoryResolver.Headers,
	}

	conn, err := cfg.Connect()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(p.ctx, p.config.ConnectionTimeout)
	defer cancel()
	if !conn.WaitForStateChange(ctx, connectivity.Ready) {
		return nil, errors.Errorf("failed to connect to remote directory %s", p.config.Addr)
	}

	return conn, nil
}

func fold(m ...api.SyncMode) api.SyncMode {
	r := api.SyncMode_SYNC_MODE_UNKNOWN
	for _, v := range m {
		r |= v
	}
	return r
}

func printMode(mode api.SyncMode) string {
	modes := []string{}
	if mode&api.SyncMode_SYNC_MODE_MANIFEST != 0 {
		modes = append(modes, "MANIFEST")
	}
	if mode&api.SyncMode_SYNC_MODE_FULL != 0 {
		modes = append(modes, "FULL")
	}
	if mode&api.SyncMode_SYNC_MODE_DIFF != 0 {
		modes = append(modes, "DIFF")
	}
	if mode&api.SyncMode_SYNC_MODE_WATERMARK != 0 {
		modes = append(modes, "WATERMARK")
	}
	if mode == api.SyncMode_SYNC_MODE_UNKNOWN {
		modes = append(modes, "UNKNOWN")
	}
	return strings.Join(modes, "|")
}

func has(mode, instance api.SyncMode) bool {
	return mode&instance != 0
}
