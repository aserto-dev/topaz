package debug

import (
	"context"
	"html/template"
	"net/http"
	"net/http/pprof"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/aserto-dev/topaz/pkg/config/handler"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

const DefaultShutdownTimeout = time.Second * 0

type Config struct {
	Enabled         bool          `json:"enabled"`
	ListenAddress   string        `json:"listen_address"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`
}

var _ = handler.Config(&Config{})

func (c *Config) SetDefaults(v *viper.Viper, p ...string) {
	v.SetDefault(strings.Join(append(p, "enabled"), "."), false)
	v.SetDefault(strings.Join(append(p, "listen_address"), "."), "0.0.0.0:6060")
	v.SetDefault(strings.Join(append(p, "shutdown_timeout"), "."), DefaultShutdownTimeout.String())
}

func (c *Config) Validate() (bool, error) {
	return true, nil
}

func (c *Config) Generate(w *os.File) error {
	tmpl, err := template.New("DEBUG").Parse(debugTemplate)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	return nil
}

const debugTemplate = `
# debug service settings.
debug:
  enabled: {{ .Enabled }}
  listen_address: '{{ .ListenAddress}}'
  shutdown_timeout: {{ .ShutdownTimeout }}
`

type Server struct {
	server *http.Server
	logger *zerolog.Logger
	cfg    *Config
}

func NewServer(cfg *Config, log *zerolog.Logger) *Server {
	if !cfg.Enabled {
		return nil
	}

	http.DefaultServeMux = http.NewServeMux()

	pprofServeMux := http.NewServeMux()

	pprofServeMux.HandleFunc("/debug/pprof/", pprof.Index)
	pprofServeMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	pprofServeMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	pprofServeMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	pprofServeMux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	pprofServeMux.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))
	pprofServeMux.Handle("/debug/pprof/block", pprof.Handler("block"))
	pprofServeMux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	pprofServeMux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	pprofServeMux.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))
	pprofServeMux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))

	debugLogger := log.With().Str("component", "debug").Logger()

	runtime.SetMutexProfileFraction(10)
	debugLogger.Info().Int("fraction", runtime.SetMutexProfileFraction(-1)).Msg("mutex profiler")

	srv := &http.Server{
		Addr:              cfg.ListenAddress,
		Handler:           pprofServeMux,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	return &Server{
		server: srv,
		logger: &debugLogger,
		cfg:    cfg,
	}
}

func (srv *Server) Start() {
	if !srv.cfg.Enabled {
		return
	}

	if srv != nil {
		go func() {
			srv.logger.Warn().Str("listen_address", srv.cfg.ListenAddress).Msg("debug-service")
			if err := srv.server.ListenAndServe(); err != nil {
				srv.logger.Error().Err(err).Msg("debug-service")
			}
		}()
	}
}

func (srv *Server) Stop() {
	if srv == nil || !srv.cfg.Enabled {
		return
	}

	var shutdown context.CancelFunc
	ctx := context.Background()
	if srv.cfg.ShutdownTimeout > 0 {
		ctx, shutdown = context.WithTimeout(ctx, srv.cfg.ShutdownTimeout)
		defer shutdown()
	}

	err := srv.server.Shutdown(ctx)
	if err != nil {
		srv.logger.Info().Err(err).Str("state", "shutdown").Msg("debug-service")
	}
}
