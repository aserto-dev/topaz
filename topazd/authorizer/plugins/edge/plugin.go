package edge

import (
	"context"
	"strings"
	"sync/atomic"
	"time"

	client "github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/topaz/internal/eds/pkg/datasync"
	"github.com/aserto-dev/topaz/internal/eds/pkg/directory"
	topaz "github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/topazd/app"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/open-policy-agent/opa/v1/plugins"
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

type SyncMode int32

const (
	// SyncModeUnknown nothing selected (default initialization value).
	SyncModeUnknown SyncMode = 0
	// SyncModeFull full sync, requests full export of source, contains new and updated elements only.
	SyncModeFull SyncMode = 1
	// SyncModeDiff full sync with differential, removing items deleted in source from target.
	SyncModeDiff SyncMode = 2
	// SyncModeWatermark watermark sync, pulls all new and updated data since last watermark.
	SyncModeWatermark SyncMode = 4
	// SyncModeManifest manifest sync, pulls manifest from source and applies to target when etags are different.
	SyncModeManifest SyncMode = 8
)

type Config struct {
	Enabled           bool              `json:"enabled"`             //
	Addr              string            `json:"addr"`                //
	APIKey            string            `json:"apikey"`              //
	Timeout           int               `json:"timeout"`             // timeout in seconds.
	SyncInterval      int               `json:"sync_interval"`       // interval in minutes.
	Insecure          bool              `json:"insecure"`            //
	ConnectionTimeout time.Duration     `json:"-"`                   // mapped at runtime to timeout * time.Second.
	PageSize          int               `json:"page_size,omitempty"` // deprecated: no longer used.
	ClientCertPath    string            `json:"client_cert_path"`    //
	ClientKeyPath     string            `json:"client_key_path"`     //
	CACertPath        string            `json:"ca_cert_path"`        //
	NoTLS             bool              `json:"no_tls"`              //
	NoProxy           bool              `json:"no_proxy"`            //
	Headers           map[string]string `json:"headers"`             //
}

type Plugin struct {
	ctx         context.Context
	cancel      context.CancelFunc
	manager     *plugins.Manager
	logger      *zerolog.Logger
	config      *Config
	topazConfig *topaz.Config
	syncNow     chan SyncMode
}

func newEdgePlugin(logger *zerolog.Logger, cfg *Config, topazConfig *topaz.Config, manager *plugins.Manager) *Plugin {
	newLogger := logger.With().Str("component", "edge.plugin").Logger()

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
		syncNow:     make(chan SyncMode),
	}
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

func (p *Plugin) Reconfigure(ctx context.Context, config any) {
	p.logger.Trace().Str("id", p.manager.ID).Interface("cur", p.config).Interface("new", config).Msg("EdgePlugin.Reconfigure")

	newConfig, ok := config.(*Config)
	if !ok {
		p.logger.Error().Str("config", "failed type assertion").Msg("EdgePlugin.Reconfigure")
		return
	}

	// handle enabled status changed
	if p.config.Enabled != newConfig.Enabled {
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

	p.config, ok = config.(*Config)
	if !ok {
		p.logger.Error().Str("config", "failed type assertion").Msg("EdgePlugin.Reconfigure")
		return
	}
}

func (p *Plugin) SyncNow(mode SyncMode) {
	p.syncNow <- mode
}

func (p *Plugin) resetContext() {
	p.ctx, p.cancel = context.WithCancel(context.Background())
}

const cycles int64 = 4

func (p *Plugin) scheduler() {
	// scheduler startup delay 1s
	interval := time.NewTicker(1 * time.Second)
	defer interval.Stop()

	var running atomic.Bool

	running.Store(false)

	cycle := cycles

	intervalMode := SyncModeUnknown
	onDemandMode := SyncModeUnknown

	for {
		select {
		case <-p.ctx.Done():
			p.logger.Debug().Time("done", time.Now()).Msg(syncScheduler)
			return

		case t := <-interval.C:
			p.logger.Info().Time("dispatch", t).Msg(syncScheduler)
			interval.Stop()

			intervalMode = SyncModeWatermark

			if cycle%cycles == 0 {
				intervalMode = SyncModeDiff
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
			runMode := SyncModeUnknown
			if onDemandMode != SyncModeUnknown {
				runMode = onDemandMode
				onDemandMode = SyncModeUnknown
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
				if onDemandMode == SyncModeUnknown {
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

const secsInMin int64 = 60

// calcInterval - calculates the next time interval in secs,
// based on the configuration SyncInterval (defined on the EdgeDirectory connection)
// returning a time.Duration.
//
// p.config.SyncInterval 1m-60m
// 1m -> 60s -> 15s interval
// 60m -> 3600s -> 900s interval.
func (p *Plugin) calcInterval() time.Duration {
	waitInSec := (int64(p.config.SyncInterval) * secsInMin) / cycles
	return time.Duration(waitInSec) * time.Second
}

func (p *Plugin) task(mode SyncMode) {
	p.logger.Info().Str(status, started).Msg(syncTask)

	defer func() {
		if r := recover(); r != nil {
			p.logger.Error().Interface("recover", r).Msg(syncTask)
		}
	}()

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

	if mode == SyncModeUnknown {
		p.exec(ctx, ds, conn, []datasync.Option{
			datasync.WithMode(datasync.Manifest),
			datasync.WithMode(datasync.Full),
		})

		return
	}

	if has(mode, SyncModeWatermark) {
		p.exec(ctx, ds, conn, []datasync.Option{
			datasync.WithMode(datasync.Manifest),
			datasync.WithMode(datasync.Watermark),
		})
	}

	if has(mode, SyncModeDiff) {
		p.exec(ctx, ds, conn, []datasync.Option{
			datasync.WithMode(datasync.Manifest),
			datasync.WithMode(datasync.Diff),
		})

		return
	}

	if has(mode, SyncModeFull) {
		p.exec(ctx, ds, conn, []datasync.Option{
			datasync.WithMode(datasync.Manifest),
			datasync.WithMode(datasync.Full),
		})

		return
	}

	if has(mode, SyncModeManifest) && !has(mode, SyncModeWatermark) {
		p.exec(ctx, ds, conn, []datasync.Option{
			datasync.WithMode(datasync.Manifest),
		})

		return
	}
}

func (p *Plugin) exec(ctx context.Context, ds *directory.Directory, conn *grpc.ClientConn, opts []datasync.Option) {
	err := ds.DataSyncClient().Sync(ctx, conn, opts...)
	if err != nil {
		p.logger.Error().Err(err).Msg(syncTask)
	}

	if p.config.Enabled && err == nil {
		app.SetServiceStatus(p.logger, "sync", grpc_health_v1.HealthCheckResponse_SERVING)
	}

	p.logger.Info().Str(status, finished).Msg(syncTask)
}

func (p *Plugin) remoteDirectoryClient() (*grpc.ClientConn, error) {
	cfg := &client.Config{
		Address:        p.config.Addr,           //
		APIKey:         p.config.APIKey,         //
		ClientCertPath: p.config.ClientCertPath, //
		ClientKeyPath:  p.config.ClientKeyPath,  //
		CACertPath:     p.config.CACertPath,     //
		Insecure:       p.config.Insecure,       //
		NoTLS:          p.config.NoTLS,          //
		NoProxy:        p.config.NoProxy,        //
		Headers:        p.config.Headers,        //
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

func fold(m ...SyncMode) SyncMode {
	r := SyncModeUnknown

	for _, v := range m {
		r |= v
	}

	return r
}

func printMode(mode SyncMode) string {
	modes := []string{}
	if mode&SyncModeManifest != 0 {
		modes = append(modes, "MANIFEST")
	}

	if mode&SyncModeFull != 0 {
		modes = append(modes, "FULL")
	}

	if mode&SyncModeDiff != 0 {
		modes = append(modes, "DIFF")
	}

	if mode&SyncModeWatermark != 0 {
		modes = append(modes, "WATERMARK")
	}

	if mode == SyncModeUnknown {
		modes = append(modes, "UNKNOWN")
	}

	return strings.Join(modes, "|")
}

func has(mode, instance SyncMode) bool {
	return mode&instance != 0
}
