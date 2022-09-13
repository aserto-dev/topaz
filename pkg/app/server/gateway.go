package server

import (
	"fmt"
	"net/http"

	promclient "github.com/prometheus/client_golang/prometheus"
	"github.com/slok/go-http-metrics/metrics/prometheus"

	"github.com/aserto-dev/aserto-grpc/grpcutil"
	metricsserver "github.com/aserto-dev/aserto-grpc/grpcutil/metrics"
	"github.com/aserto-dev/go-utils/certs"
	"github.com/aserto-dev/go-utils/logger"
	openapi "github.com/aserto-dev/openapi-grpc/publish/authorizer"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/grpc"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	allowedOrigins = []string{
		"http://localhost",
		"http://localhost:*",
		"https://localhost",
		"https://localhost:*",
		"http://127.0.0.1",
		"http://127.0.0.1:*",
		"https://127.0.0.1",
		"https://127.0.0.1:*",
	}
)

// NewGatewayServer creates a new gateway server.
func NewGatewayServer(
	log *zerolog.Logger,
	cfg *config.Common,
	gtwMux *runtime.ServeMux,
	routeRegistrations HTTPRouteRegistrations,
	registry promclient.Registerer,
) (*http.Server, error) {
	c := cors.New(cors.Options{
		AllowedHeaders: []string{"Authorization", "Content-Type", "Depth", string(grpcutil.HeaderAsertoTenantID),
			string(grpcutil.HeaderAsertoTenantKey)},
		AllowedOrigins: append(allowedOrigins, cfg.API.Gateway.AllowedOrigins...),
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodHead, http.MethodDelete, http.MethodPut,
			http.MethodPatch, "PROPFIND", "MKCOL", "COPY", "MOVE"},
		Debug: cfg.Logging.LogLevelParsed <= zerolog.DebugLevel,
	})
	c.Log = log

	newLogger := log.With().Str("source", "http-gateway").Logger()

	metricsmw := metricsMiddleware(&cfg.API.Metrics, registry)

	mux := http.NewServeMux()
	mux.Handle("/api/", metricsmw(fieldsMaskHandler(gtwMux)))
	mux.Handle("/openapi.json", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		http.FileServer(http.FS(openapi.Static())).ServeHTTP(w, r)
	}))
	mux.Handle("/robots.txt", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "User-agent: *\nDisallow: /")
	}))
	routeRegistrations(mux)

	gtwServer := &http.Server{
		ErrorLog: logger.NewSTDLogger(&newLogger),
		Addr:     cfg.API.Gateway.ListenAddress,
		Handler:  c.Handler(mux),
	}

	if cfg.API.Gateway.HTTP {
		return gtwServer, nil
	}

	tlsServerConfig, err := certs.GatewayServerTLSConfig(cfg.API.Gateway.Certs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to calculate gateway server tls creds")
	}

	gtwServer.TLSConfig = tlsServerConfig

	return gtwServer, nil
}

func metricsMiddleware(cfg *metricsserver.Config, registry promclient.Registerer) func(http.Handler) http.Handler {
	if !cfg.GRPC.Gateway {
		return func(h http.Handler) http.Handler { return h }
	}

	return grpc.GatewayMuxMetricsMiddleware(
		middleware.New(middleware.Config{
			Recorder: prometheus.NewRecorder(prometheus.Config{Registry: registry}),
		}),
	)
}

// fieldsMaskHandler if a fields.mask query parameter is present and set,
// the handler will set the Content-Type to "application/json+masked", which
// will signal the marshaler to not emit unpopulated types, which is needed to
// serialize the masked result set.
func fieldsMaskHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p, ok := r.URL.Query()["fields.mask"]; ok && len(p) > 0 && len(p[0]) > 0 {
			r.Header.Set("Content-Type", "application/json+masked")
		}
		h.ServeHTTP(w, r)
	})
}

// customHeaderMatcher so that HTTP clients do not have to prefix the header key with Grpc-Metadata-
// see https://grpc-ecosystem.github.io/grpc-gateway/docs/mapping/customizing_your_gateway/#mapping-from-http-request-headers-to-grpc-client-metadata
func customHeaderMatcher(key string) (string, bool) {
	switch key {
	case string(grpcutil.HeaderAsertoTenantKey):
		return key, true
	case string(grpcutil.HeaderAsertoTenantID):
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

// GatewayMux creates a gateway multiplexer for serving the API as an OpenAPI endpoint.
func GatewayMux() *runtime.ServeMux {
	return runtime.NewServeMux(
		runtime.WithMetadata(grpc.CaptureGatewayRoute),
		runtime.WithIncomingHeaderMatcher(customHeaderMatcher),
		runtime.WithMarshalerOption(
			runtime.MIMEWildcard,
			&runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					Multiline:       false,
					Indent:          "  ",
					AllowPartial:    true,
					UseProtoNames:   true,
					UseEnumNumbers:  false,
					EmitUnpopulated: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					AllowPartial:   true,
					DiscardUnknown: true,
				},
			},
		),
		runtime.WithErrorHandler(grpcutil.CustomErrorHandler),
		runtime.WithMarshalerOption(
			"application/json+masked",
			&runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					Multiline:       false,
					Indent:          "  ",
					AllowPartial:    true,
					UseProtoNames:   true,
					UseEnumNumbers:  false,
					EmitUnpopulated: false,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					AllowPartial:   true,
					DiscardUnknown: true,
				},
			},
		),
	)
}
