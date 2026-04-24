package openapi

import (
	_ "embed"
	"net/http"

	"github.com/rs/zerolog"
)

type Server struct {
	Scheme string
	Host   string
}

//go:embed openapi.json
var static []byte

// Static, serve the openapi.json file.
func Static() []byte {
	return static
}

var cache syncMap[string, *templateBuilder]

// OpenAPIHandler, handler to serve the OpenAPI specification file.
func OpenAPIHandler(port string, svc ...string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		builder, loaded := cache.LoadOrStore(port, newTemplateBuilder())
		if !loaded {
			// First time being called with this port.
			// Build the template
			builder.Build(port, svc...)
		}

		tmpl, err := builder.Get()
		if err != nil {
			zerolog.Ctx(r.Context()).Err(err).Msg("failed to build template")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)

			return
		}

		server := Server{
			Host:   r.Host,
			Scheme: scheme(r),
		}

		if err := tmpl.Execute(w, server); err != nil {
			zerolog.Ctx(r.Context()).Err(err).Msg("failed to execute template")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)

			return
		}
	}
}

func scheme(r *http.Request) string {
	scheme := r.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		if r.TLS == nil {
			scheme = "http"
		} else {
			scheme = "https"
		}
	}

	return scheme
}
