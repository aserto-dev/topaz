// Package debug implements an HTTP server that exposes pprof endpoints for debugging and profiling Go applications.
package debug

import (
	"context"
	"html/template"
	"io"
	"net/http"
	"net/http/pprof"
	"runtime"

	gorilla "github.com/gorilla/mux"
	"github.com/rs/zerolog"

	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/topazd/servers"
)

type Config struct {
	config.Optional
	servers.HTTPServer
}

var _ config.Section = (*Config)(nil)

func (c *Config) Defaults() map[string]any {
	return map[string]any{
		"enabled":        false,
		"listen_address": "0.0.0.0:6060",
	}
}

func (c *Config) Validate() error {
	return nil
}

func (c *Config) Serialize(w io.Writer) error {
	tmpl, err := template.New("DEBUG").Parse(config.TrimN(debugTemplate))
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	return nil
}

func RegisterHandlers(ctx context.Context, router *gorilla.Router) {
	router.HandleFunc("/pprof/", pprof.Index).Methods(http.MethodGet)
	router.HandleFunc("/pprof/cmdline", pprof.Cmdline).Methods(http.MethodGet)
	router.HandleFunc("/pprof/profile", pprof.Profile).Methods(http.MethodGet)
	router.HandleFunc("/pprof/symbol", pprof.Symbol).Methods(http.MethodGet)
	router.HandleFunc("/pprof/trace", pprof.Trace).Methods(http.MethodGet)

	router.Handle("/pprof/allocs", pprof.Handler("allocs")).Methods(http.MethodGet)
	router.Handle("/pprof/block", pprof.Handler("block")).Methods(http.MethodGet)
	router.Handle("/pprof/goroutine", pprof.Handler("goroutine")).Methods(http.MethodGet)
	router.Handle("/pprof/heap", pprof.Handler("heap")).Methods(http.MethodGet)
	router.Handle("/pprof/mutex", pprof.Handler("mutex")).Methods(http.MethodGet)
	router.Handle("/pprof/threadcreate", pprof.Handler("threadcreate")).Methods(http.MethodGet)

	zerolog.Ctx(ctx).Info().Int("fraction", runtime.SetMutexProfileFraction(-1)).Msg("mutex profiler")
}

const debugTemplate = `
# debug service settings.
debug:
  enabled: {{ .Enabled }}
{{- with .ListenAddress }}
  listen_address: '{{ . }}'
{{- end }}
`
