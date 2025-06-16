package debug

import (
	"context"
	"html/template"
	"io"
	"net/http/pprof"
	"runtime"

	gorilla "github.com/gorilla/mux"
	"github.com/rs/zerolog"

	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/pkg/servers"
)

type Config struct {
	servers.HTTPServer `json:",squash"` //nolint:staticcheck,tagliatelle  // squash is part of mapstructure

	Enabled bool `json:"enabled"`
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
	router.HandleFunc("/pprof/", pprof.Index)
	router.HandleFunc("/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/pprof/profile", pprof.Profile)
	router.HandleFunc("/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/pprof/trace", pprof.Trace)

	router.Handle("/pprof/allocs", pprof.Handler("allocs"))
	router.Handle("/pprof/block", pprof.Handler("block"))
	router.Handle("/pprof/goroutine", pprof.Handler("goroutine"))
	router.Handle("/pprof/heap", pprof.Handler("heap"))
	router.Handle("/pprof/mutex", pprof.Handler("mutex"))
	router.Handle("/pprof/threadcreate", pprof.Handler("threadcreate"))

	zerolog.Ctx(ctx).Info().Int("fraction", runtime.SetMutexProfileFraction(-1)).Msg("mutex profiler")
}

const debugTemplate = `
# debug service settings.
debug:
  enabled: {{ .Enabled }}
  listen_address: '{{ .ListenAddress}}'
`
