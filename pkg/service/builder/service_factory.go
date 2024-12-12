package builder

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/aserto-dev/go-aserto"
	"github.com/samber/lo"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/cors"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type ServiceFactory struct{}

var mdlw middleware.Middleware

func NewServiceFactory() *ServiceFactory {
	if reg == nil {
		reg = prometheus.NewRegistry()
	}
	mdlw = middleware.New(middleware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{
			Registry: reg,
		}),
	})

	return &ServiceFactory{}
}

type GRPCOptions struct {
	ServerOptions []grpc.ServerOption
	Registrations GRPCRegistrations
}

type GatewayOptions struct {
	HandlerRegistrations HandlerRegistrations
	ErrorHandler         runtime.ErrorHandlerFunc
}

func (f *ServiceFactory) CreateService(
	config *API,
	grpcOpts *GRPCOptions,
	gatewayOpts *GatewayOptions,
	cleanup ...func(),
) (*Service, error) {
	grpcServer, err := prepareGrpcServer(&config.GRPC.Certs, grpcOpts.ServerOptions)
	if err != nil {
		return nil, err
	}
	grpcOpts.Registrations(grpcServer)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPC.ListenAddress)
	if err != nil {
		return nil, err
	}

	var gateway *Gateway
	if gatewayOpts != nil && config.Gateway.ListenAddress != "" {
		gateway, err = f.prepareGateway(config, gatewayOpts)
		if err != nil {
			return nil, err
		}
	}

	return &Service{
		Config:   config,
		Server:   grpcServer,
		Listener: listener,
		Gateway:  gateway,
		Started:  make(chan bool),
		Cleanup:  cleanup,
	}, nil
}

// prepareGateway provides a http server that will have the registrations pointed to the corresponding configured grpc server.
func (f *ServiceFactory) prepareGateway(config *API, gatewayOpts *GatewayOptions) (*Gateway, error) {
	if len(config.Gateway.AllowedHeaders) == 0 {
		config.Gateway.AllowedHeaders = DefaultGatewayAllowedHeaders
	}
	if len(config.Gateway.AllowedOrigins) == 0 {
		config.Gateway.AllowedOrigins = DefaultGatewayAllowedOrigins
	}
	if len(config.Gateway.AllowedMethods) == 0 {
		config.Gateway.AllowedMethods = DefaultGatewayAllowedMethods
	}

	c := cors.New(cors.Options{
		AllowedHeaders: config.Gateway.AllowedHeaders,
		AllowedOrigins: config.Gateway.AllowedOrigins,
		AllowedMethods: config.Gateway.AllowedMethods,
		Debug:          false,
	})

	runtimeMux := f.gatewayMux(config.Gateway.AllowedHeaders, gatewayOpts.ErrorHandler)

	opts := []grpc.DialOption{}
	if config.GRPC.Certs.HasCA() {
		tlsCreds, err := config.GRPC.Certs.ClientCredentials(true)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get TLS credentials")
		}

		opts = append(opts, grpc.WithTransportCredentials(tlsCreds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	grpcEndpoint := fmt.Sprintf("dns:///%s", config.GRPC.ListenAddress)

	if err := gatewayOpts.HandlerRegistrations(context.Background(), runtimeMux, grpcEndpoint, opts); err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.Handle("/", runtimeMux)
	mux.Handle("/api/", fieldsMaskHandler(runtimeMux))

	gtwHandler := std.Handler("", mdlw, mux)

	gtwServer := &http.Server{
		Addr:              config.Gateway.ListenAddress,
		Handler:           c.Handler(gtwHandler),
		ReadTimeout:       config.Gateway.ReadTimeout,
		ReadHeaderTimeout: config.Gateway.ReadHeaderTimeout,
		WriteTimeout:      config.Gateway.WriteTimeout,
		IdleTimeout:       config.Gateway.IdleTimeout,
	}

	config.Gateway.HTTP = !config.Gateway.Certs.HasCert()

	if config.Gateway.HTTP {
		return &Gateway{Server: gtwServer, Mux: mux, Certs: nil}, nil
	}

	tlsServerConfig, err := config.Gateway.Certs.ServerConfig()
	if err != nil {
		return nil, err
	}

	gtwServer.TLSConfig = tlsServerConfig
	return &Gateway{Server: gtwServer, Mux: mux, Certs: &config.Gateway.Certs}, nil
}

// gatewayMux creates a gateway multiplexer for serving the API as an OpenAPI endpoint.
func (f *ServiceFactory) gatewayMux(allowedHeaders []string, errorHandler runtime.ErrorHandlerFunc) *runtime.ServeMux {
	opts := []runtime.ServeMuxOption{
		runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
			if lo.Contains(allowedHeaders, key) {
				return key, true
			}
			return runtime.DefaultHeaderMatcher(key)
		}),
		runtime.WithMetadata(captureGatewayRoute),
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
		runtime.WithUnescapingMode(runtime.UnescapingModeAllExceptSlash),
		runtime.WithForwardResponseOption(httpResponseModifier),
	}

	if errorHandler != nil {
		opts = append(opts, runtime.WithErrorHandler(errorHandler))
	}

	return runtime.NewServeMux(opts...)
}

func httpResponseModifier(ctx context.Context, w http.ResponseWriter, p proto.Message) error {
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		return nil
	}

	// set http status code
	if vals := md.HeaderMD.Get("x-http-code"); len(vals) > 0 {
		code, err := strconv.Atoi(vals[0])
		if err != nil {
			return err
		}
		// delete the headers to not expose any grpc-metadata in http response
		delete(md.HeaderMD, "x-http-code")
		delete(w.Header(), "Grpc-Metadata-X-Http-Code")
		w.WriteHeader(code)
	}

	return nil
}

// prepareGrpcServer provides a new grpc server with the provided grpc.ServerOptions using the provided certificates.
func prepareGrpcServer(certCfg *aserto.TLSConfig, opts []grpc.ServerOption) (*grpc.Server, error) {
	// NoTLS path.
	if !certCfg.HasCert() {
		opts = append(opts, grpc.Creds(insecure.NewCredentials()))
		return grpc.NewServer(opts...), nil
	}

	// TLS path.
	tlsCreds, err := certCfg.ServerCredentials()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get TLS credentials")
	}

	tlsAuth := grpc.Creds(tlsCreds)
	opts = append(opts, tlsAuth)

	return grpc.NewServer(opts...), nil
}

// fieldsMaskHandler will set the Content-Type to "application/json+masked", which
// will signal the marshaler to not emit unpopulated types, which is needed to
// serialize the masked result set.
// This happens if a fields.mask query parameter is present and set.
func fieldsMaskHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p, ok := r.URL.Query()["fields.mask"]; ok && len(p) > 0 && p[0] != "" {
			r.Header.Set("Content-Type", "application/json+masked")
		}
		h.ServeHTTP(w, r)
	})
}

type key int

var pathPatternKey key

type gatewayPathPattern struct {
	PathPattern string
}

func captureGatewayRoute(ctx context.Context, r *http.Request) metadata.MD {
	if pattern, ok := runtime.HTTPPathPattern(ctx); ok {
		if gwPathPattern := gatewayContextValue(r); gwPathPattern != nil {
			gwPathPattern.PathPattern = pattern
		}
	}
	return nil
}

func gatewayContextValue(r *http.Request) *gatewayPathPattern {
	gwPathPattern, ok := r.Context().Value(pathPatternKey).(*gatewayPathPattern)
	if !ok {
		return nil
	}

	return gwPathPattern
}
