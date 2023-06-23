package builder

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/aserto-dev/certs"
	"github.com/aserto-dev/go-edge-ds/pkg/directory"
	metrics "github.com/aserto-dev/go-http-metrics/middleware/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

type ServiceFactory struct{}

func NewServiceFactory() *ServiceFactory {
	return &ServiceFactory{}
}

func (f *ServiceFactory) CreateService(config *directory.API, opts []grpc.ServerOption, registrations GRPCRegistrations,
	handlerRegistrations HandlerRegistrations,
	withGateway bool) (*Server, error) {
	grpcServer, err := prepareGrpcServer(&config.GRPC.Certs, opts)
	if err != nil {
		return nil, err
	}
	registrations(grpcServer)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPC.ListenAddress)
	if err != nil {
		return nil, err
	}
	gate := Gateway{}
	if withGateway && config.Gateway.ListenAddress != "" {
		gate, err = prepareGateway(config, handlerRegistrations)
		if err != nil {
			return nil, err
		}
	}

	return &Server{
		Config:        config,
		Server:        grpcServer,
		Listener:      listener,
		Registrations: registrations,
		Gateway:       gate,
	}, nil
}

var allowedOrigins = []string{
	"http://localhost",
	"http://localhost:*",
	"https://localhost",
	"https://localhost:*",
	"http://127.0.0.1",
	"http://127.0.0.1:*",
	"https://127.0.0.1",
	"https://127.0.0.1:*",
}

func prepareGateway(config *directory.API, registrations HandlerRegistrations) (Gateway, error) {
	c := cors.New(cors.Options{
		AllowedHeaders: []string{"Authorization", "Content-Type", "Depth"},
		AllowedOrigins: append(allowedOrigins, config.Gateway.AllowedOrigins...),
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodHead, http.MethodDelete, http.MethodPut,
			http.MethodPatch, "PROPFIND", "MKCOL", "COPY", "MOVE"},
		Debug: true,
	})

	runtimeMux := gatewayMux()

	tlsCreds, err := certs.GatewayAsClientTLSCreds(config.GRPC.Certs)
	if err != nil {
		return Gateway{}, errors.Wrapf(err, "failed to get TLS credentials")
	}
	grpcEndpoint := fmt.Sprintf("dns:///%s", config.GRPC.ListenAddress)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(tlsCreds),
	}

	err = registrations(context.Background(), runtimeMux, grpcEndpoint, opts)
	if err != nil {
		return Gateway{}, err
	}

	mux := http.NewServeMux()
	mux.Handle("/", runtimeMux)

	gtwServer := &http.Server{
		Addr:              config.Gateway.ListenAddress,
		Handler:           c.Handler(mux),
		ReadTimeout:       config.Gateway.ReadTimeout,
		ReadHeaderTimeout: config.Gateway.ReadHeaderTimeout,
		WriteTimeout:      config.Gateway.WriteTimeout,
		IdleTimeout:       config.Gateway.IdleTimeout,
	}

	return Gateway{Server: gtwServer, Certs: &config.Gateway.Certs}, nil
}

// gatewayMux creates a gateway multiplexer for serving the API as an OpenAPI endpoint.
func gatewayMux() *runtime.ServeMux {
	return runtime.NewServeMux(
		runtime.WithMetadata(metrics.CaptureGatewayRoute),
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
		// runtime.WithErrorHandler(CustomErrorHandler),
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

func prepareGrpcServer(certCfg *certs.TLSCredsConfig, opts []grpc.ServerOption) (*grpc.Server, error) {
	if certCfg != nil && certCfg.TLSCertPath != "" {
		tlsCreds, err := certs.GRPCServerTLSCreds(*certCfg)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get TLS credentials")
		}
		tlsAuth := grpc.Creds(tlsCreds)
		opts = append(opts, tlsAuth)
	}
	return grpc.NewServer(opts...), nil
}
