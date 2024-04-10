package debug

import (
	"context"
	"net/http"
	"net/http/pprof"
	_ "net/http/pprof"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

type Config struct {
	Enabled         bool   `json:"enabled"`
	ListenAddress   string `json:"listen_address"`
	ShutdownTimeout int    `json:"shutdown_timeout"`
}

type Server struct {
	server   *http.Server
	logger   *zerolog.Logger
	cfg      *Config
	errGroup *errgroup.Group
}

func NewServer(cfg *Config, log *zerolog.Logger, errGroup *errgroup.Group) *Server {
	if cfg.Enabled {
		pprofMux := http.NewServeMux()
		pprofMux.Handle("/debug/allocs", pprof.Handler("allocs"))
		pprofMux.Handle("/debug/block", pprof.Handler("block"))
		pprofMux.Handle("/debug/goroutine", pprof.Handler("goroutine"))
		pprofMux.Handle("/debug/heap", pprof.Handler("heap"))
		pprofMux.Handle("/debug/mutex", pprof.Handler("mutex"))
		pprofMux.Handle("/debug/threadcreate", pprof.Handler("threadcreate"))
		pprofMux.Handle("/debug/profile", http.HandlerFunc(pprof.Profile))
		pprofMux.Handle("/debug/symbol", http.HandlerFunc(pprof.Symbol))
		pprofMux.Handle("/debug/trace", http.HandlerFunc(pprof.Trace))

		srv := &http.Server{
			Addr:              cfg.ListenAddress,
			Handler:           pprofMux,
			ReadTimeout:       5 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       30 * time.Second,
		}

		debugLogger := log.With().Str("component", "debug").Logger()

		return &Server{
			server:   srv,
			logger:   &debugLogger,
			cfg:      cfg,
			errGroup: errGroup,
		}
	}

	return nil
}

func (srv *Server) Start() {
	if srv != nil {
		srv.errGroup.Go(func() error {
			err := srv.server.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				srv.logger.Error().Err(err).Str("address", srv.cfg.ListenAddress).Msg("Profiling endpoint failed to listen")
			}
			return nil
		})
	}
}

func (srv *Server) Stop() {
	if srv != nil {
		var shutdown context.CancelFunc
		ctx := context.Background()
		if srv.cfg.ShutdownTimeout > 0 {
			shutdownTimeout := time.Duration(srv.cfg.ShutdownTimeout) * time.Second
			ctx, shutdown = context.WithTimeout(ctx, shutdownTimeout)
			defer shutdown()
		}
		err := srv.server.Shutdown(ctx)
		if err != nil {
			srv.logger.Info().Err(err).Msg("error shutting down debug server")
		}
	}
}
