package server

import (
	"fmt"
	"net/http"

	promclient "github.com/prometheus/client_golang/prometheus"

	"github.com/aserto-dev/certs"
	"github.com/aserto-dev/logger"
	openapi "github.com/aserto-dev/openapi-grpc/publish/authorizer"
	"github.com/aserto-dev/topaz/pkg/cc/config"
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
	registry promclient.Registerer,
) (*http.Server, error) {
	c := cors.New(cors.Options{
		AllowedHeaders: []string{"Authorization", "Content-Type", "Depth"},
		AllowedOrigins: append(allowedOrigins, cfg.API.Gateway.AllowedOrigins...),
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodHead, http.MethodDelete, http.MethodPut,
			http.MethodPatch, "PROPFIND", "MKCOL", "COPY", "MOVE"},
		Debug: cfg.Logging.LogLevelParsed <= zerolog.DebugLevel,
	})
	c.Log = log

	newLogger := log.With().Str("source", "http-gateway").Logger()

	mux := http.NewServeMux()
	mux.Handle("/openapi.json", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		http.FileServer(http.FS(openapi.Static())).ServeHTTP(w, r)
	}))
	mux.Handle("/robots.txt", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "User-agent: *\nDisallow: /")
	}))

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

// GatewayMux creates a gateway multiplexer for serving the API as an OpenAPI endpoint.
func GatewayMux() *runtime.ServeMux {
	return runtime.NewServeMux(
		runtime.WithMetadata(grpc.CaptureGatewayRoute),
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
		// TODO: figure out if we need a custom error handler or not
		// runtime.WithErrorHandler(grpcutil.CustomErrorHandler),
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
