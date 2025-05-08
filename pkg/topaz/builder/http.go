package builder

import (
	"context"
	"errors"
	"net/http"

	"github.com/aserto-dev/topaz/pkg/servers"
	gorilla "github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog"
)

var noHTTP = new(httpServer)

type httpServer struct {
	*http.Server

	router *gorilla.Router
}

func newHTTPServer(cfg *servers.HTTPServer) (*httpServer, error) {
	router := gorilla.NewRouter()

	tlsConf, err := cfg.Certs.ServerConfig()
	if err != nil {
		return nil, err
	}

	return &httpServer{
		Server: &http.Server{
			Addr:              cfg.ListenAddress,
			TLSConfig:         tlsConf,
			Handler:           cfg.Cors().Handler(router),
			ReadTimeout:       cfg.ReadTimeout,
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
			WriteTimeout:      cfg.WriteTimeout,
			IdleTimeout:       cfg.IdleTimeout,
		},
		router: router,
	}, nil
}

func (s *httpServer) Start(ctx context.Context, runner Runner) error {
	if !s.Enabled() {
		// No http server.
		return nil
	}

	runner.Go(func() error {
		var err error

		if s.TLSConfig == nil {
			err = s.ListenAndServe()
		} else {
			// Certs are already provided in the server's TLSConfig.
			err = s.ListenAndServeTLS("", "")
		}

		if errors.Is(err, http.ErrServerClosed) {
			// ErrServerClosed is returned after normal Shutdown or Close.
			err = nil
		}

		return err
	})

	return nil
}

func (s *httpServer) Stop(ctx context.Context) error {
	if err := s.Shutdown(ctx); err != nil {
		zerolog.Ctx(ctx).Err(err).Msg("failed to shutdown http server")

		return s.Close()
	}

	return nil
}

func (s *httpServer) Enabled() bool {
	return s.Server != nil
}

func (s *httpServer) AttachGateway(prefix string, mux *runtime.ServeMux) {
	apiRouter := s.router.PathPrefix(prefix).Subrouter()
	apiRouter.Use(FieldsMask)
	apiRouter.PathPrefix("/").Handler(mux)
}

// FieldsMask is http middlware that sets the Content-Type to "application/json+masked", which
// signals the marshaler not to emit unpopulated types, which is needed to
// serialize the masked result set.
// This happens if a fields.mask query parameter is present and set.
func FieldsMask(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p, ok := r.URL.Query()["fields.mask"]; ok && len(p) > 0 && p[0] != "" {
			r.Header.Set("Content-Type", "application/json+masked")
		}

		h.ServeHTTP(w, r)
	})
}
