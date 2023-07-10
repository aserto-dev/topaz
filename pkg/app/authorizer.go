package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/aserto-dev/certs"
	"github.com/aserto-dev/go-directory/aserto/directory/exporter/v2"
	"github.com/aserto-dev/go-directory/aserto/directory/importer/v2"
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v2"
	"github.com/aserto-dev/go-directory/aserto/directory/writer/v2"
	"github.com/aserto-dev/go-edge-ds/pkg/directory"
	"github.com/aserto-dev/go-edge-ds/pkg/session"
	"github.com/aserto-dev/topaz/pkg/app/impl"

	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"google.golang.org/grpc"

	edge "github.com/aserto-dev/go-edge-ds/pkg/server"

	authz "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	openapi "github.com/aserto-dev/openapi-authorizer/publish/authorizer"

	diropenapi "github.com/aserto-dev/openapi-directory/publish/directory"
	builder "github.com/aserto-dev/service-host"
)

var locker edge.EdgeDirLock

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

// Authorizer is an authorizer service instance, responsible for managing
// the authorizer API, user directory instance and the OPA plugins.
type Authorizer struct {
	Context          context.Context
	Logger           *zerolog.Logger
	ServerOptions    []grpc.ServerOption
	Configuration    *config.Config
	ServiceBuilder   *builder.ServiceFactory
	Manager          *builder.ServiceManager
	Resolver         *resolvers.Resolvers
	AuthorizerServer *impl.AuthorizerServer
}

func (e *Authorizer) AddGRPCServerOptions(grpcOptions ...grpc.ServerOption) {
	e.ServerOptions = append(e.ServerOptions, grpcOptions...)
}

// Start starts all services required by the engine.
func (e *Authorizer) Start() error {
	err := e.configServices()
	if err != nil {
		return err
	}

	err = e.Manager.StartServers(e.Context)
	if err != nil {
		return errors.Wrap(err, "failed to start engine server")
	}

	return nil
}

func (e *Authorizer) configServices() error {
	if readerConfig, ok := e.Configuration.Services["reader"]; ok {
		if readerConfig.GRPC.ListenAddress != e.Configuration.DirectoryResolver.Address {
			return errors.New("remote directory resolver address is different from reader grpc address")
		}
	}
	serviceMap := mapToGRPCPorts(e.Configuration.Services)

	if cfg, ok := e.Configuration.Services["authorizer"]; ok {
		// attach default allowed origins to gateway.
		cfg.Gateway.AllowedOrigins = append(cfg.Gateway.AllowedOrigins, allowedOrigins...)
		server, err := e.configAuthorizer(cfg)
		if err != nil {
			return err
		}

		err = e.Manager.AddGRPCServer(server)
		if err != nil {
			return err
		}

		if len(serviceMap) == 1 &&
			e.Configuration.DirectoryResolver.Address != e.Configuration.Services["authorizer"].GRPC.ListenAddress &&
			isLocalDirectory(e.Configuration.DirectoryResolver.Address) {
			defaultAPI := builder.API{}
			defaultAPI.GRPC.ListenAddress = e.Configuration.DirectoryResolver.Address
			defaultAPI.GRPC.Certs = e.Configuration.Services["authorizer"].GRPC.Certs
			serviceMap[e.Configuration.DirectoryResolver.Address] = services{
				registeredServices: []string{"reader", "writer", "importer", "exporter"},
				API:                &defaultAPI,
			}
		}
	}

	for _, config := range serviceMap {
		if contains(config.registeredServices, "authorizer") {
			continue
		}
		serviceConfig := config
		server, err := e.configEdgeDir(&serviceConfig)
		if err != nil {
			return err
		}
		err = e.Manager.AddGRPCServer(server)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Authorizer) configEdgeDir(cfg *services) (*builder.Server, error) {
	if cfg.API.GRPC.Certs.TLSCACertPath == "" {
		if _, ok := e.Configuration.Services["authorizer"]; ok {
			cfg.API.GRPC.Certs = e.Configuration.Services["authorizer"].GRPC.Certs
		} else {
			return nil, errors.New("GRPC certificates required for edge services")
		}
	}
	if cfg.API.Gateway.Certs.TLSCACertPath == "" {
		cfg.API.Gateway.HTTP = true
	}

	if len(cfg.API.Needs) > 0 {
		for _, name := range cfg.API.Needs {
			if dependencyConfig, ok := e.Configuration.Services[name]; ok {
				if !contains(e.Manager.DependencyMap[cfg.API.GRPC.ListenAddress], dependencyConfig.GRPC.ListenAddress) {
					e.Manager.DependencyMap[cfg.API.GRPC.ListenAddress] = append(e.Manager.DependencyMap[cfg.API.GRPC.ListenAddress], dependencyConfig.GRPC.ListenAddress)
				}
			}
		}
	}

	// attach default allowed origins to gateway.
	cfg.API.Gateway.AllowedOrigins = append(cfg.API.Gateway.AllowedOrigins, allowedOrigins...)

	sessionMiddleware := session.HeaderMiddleware{DisableValidation: false}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(sessionMiddleware.Unary()),
		grpc.StreamInterceptor(sessionMiddleware.Stream()),
	}
	edgeDir, err := locker.New(&e.Configuration.Edge, e.Logger)
	if err != nil {
		return nil, err
	}
	server, err := e.ServiceBuilder.CreateService(cfg.API, opts,
		e.getEdgeRegistrations(cfg.registeredServices, edgeDir),
		e.getEdgeGatewayRegistration(cfg.registeredServices), true)
	if err != nil {
		return nil, err
	}

	// attach handler for directory openapi spec.
	if server.Gateway.Mux != nil {
		server.Gateway.Mux.Handle("/api/v2/directory/openapi.json", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			http.FileServer(http.FS(diropenapi.Static())).ServeHTTP(w, r)
		}))
	}

	return server, nil
}

func (e *Authorizer) configAuthorizer(cfg *builder.API) (*builder.Server, error) {
	authorizerOpts := e.ServerOptions

	if cfg.GRPC.Certs.TLSCertPath != "" {
		tlsCreds, err := certs.GRPCServerTLSCreds(cfg.GRPC.Certs)
		if err != nil {
			return nil, errors.Wrap(err, "failed to calculate tls config")
		}

		tlsAuth := grpc.Creds(tlsCreds)
		authorizerOpts = append(authorizerOpts, tlsAuth)
	}
	if len(cfg.Needs) > 0 {
		for _, name := range cfg.Needs {
			if dependencyConfig, ok := e.Configuration.Services[name]; ok {
				if !contains(e.Manager.DependencyMap[cfg.GRPC.ListenAddress], dependencyConfig.GRPC.ListenAddress) {
					e.Manager.DependencyMap[cfg.GRPC.ListenAddress] = append(e.Manager.DependencyMap[cfg.GRPC.ListenAddress], dependencyConfig.GRPC.ListenAddress)
				}
			}
		}
	}
	// TODO: debug this - having issues with gateway connectivity when connection timeout is set
	// authorizerOpts = append(authorizerOpts, grpc.ConnectionTimeout(time.Duration(config.GRPC.ConnectionTimeoutSeconds)))

	if err := view.Register(ocgrpc.DefaultServerViews...); err != nil {
		return nil, err
	}
	authorizerOpts = append(authorizerOpts, grpc.StatsHandler(&ocgrpc.ServerHandler{}))

	var server *builder.Server
	var err error

	if e.Configuration.DirectoryResolver.Address == e.Configuration.Services["authorizer"].GRPC.ListenAddress {
		server, err = e.createAuhtorizerWithEdgeRegistrations(cfg, authorizerOpts)
		if err != nil {
			return nil, err
		}
	} else {
		server, err = e.createAuthorizer(cfg, authorizerOpts)
		if err != nil {
			return nil, err
		}
	}

	if server.Gateway.Mux != nil {
		server.Gateway.Mux.Handle("/robots.txt", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "User-agent: *\nDisallow: /")
		}))
	}
	return server, nil
}

type services struct {
	registeredServices []string
	API                *builder.API
}

func mapToGRPCPorts(api map[string]*builder.API) map[string]services {
	portMap := make(map[string]services)
	for key, config := range api {
		serv := services{}
		if existing, ok := portMap[config.GRPC.ListenAddress]; ok {
			serv = existing
			serv.registeredServices = append(serv.registeredServices, key)
		} else {
			serv.registeredServices = append(serv.registeredServices, key)
			serv.API = config
		}
		portMap[config.GRPC.ListenAddress] = serv
	}
	return portMap
}

func contains[T comparable](slice []T, item T) bool {
	for i := range slice {
		if slice[i] == item {
			return true
		}
	}
	return false
}

func (e *Authorizer) getAuthorizerRegistration() builder.GRPCRegistrations {
	return func(server *grpc.Server) {
		authz.RegisterAuthorizerServer(server, e.AuthorizerServer)
	}
}

func (e *Authorizer) getAuthorizerGatewayRegistrations() builder.HandlerRegistrations {
	return authz.RegisterAuthorizerHandlerFromEndpoint
}

func (e *Authorizer) getEdgeRegistrations(registeredServices []string, edgeDir *directory.Directory) builder.GRPCRegistrations {
	return func(server *grpc.Server) {
		if contains(registeredServices, "reader") {
			reader.RegisterReaderServer(server, edgeDir)
		}
		if contains(registeredServices, "writer") {
			writer.RegisterWriterServer(server, edgeDir)
		}
		if contains(registeredServices, "importer") {
			importer.RegisterImporterServer(server, edgeDir)
		}
		if contains(registeredServices, "exporter") {
			exporter.RegisterExporterServer(server, edgeDir)
		}
	}
}

func (e *Authorizer) getEdgeGatewayRegistration(registeredServices []string) builder.HandlerRegistrations {
	return func(ctx context.Context, mux *runtime.ServeMux, grpcEndpoint string, opts []grpc.DialOption) error {
		if contains(registeredServices, "reader") {
			err := reader.RegisterReaderHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
			if err != nil {
				return err
			}
		}
		if contains(registeredServices, "writer") {
			err := writer.RegisterWriterHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
			if err != nil {
				return err
			}
		}
		if contains(registeredServices, "importer") {
			err := importer.RegisterImporterHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
			if err != nil {
				return err
			}
		}
		if contains(registeredServices, "exporter") {
			err := exporter.RegisterExporterHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func (e *Authorizer) createAuhtorizerWithEdgeRegistrations(cfg *builder.API, authorizerOpts []grpc.ServerOption) (*builder.Server, error) {
	edgeDir, err := locker.New(&e.Configuration.Edge, e.Logger)
	if err != nil {
		return nil, err
	}
	server, err := e.ServiceBuilder.CreateService(cfg, authorizerOpts,
		func(server *grpc.Server) {
			e.getAuthorizerRegistration()(server)
			e.getEdgeRegistrations([]string{"reader", "writer", "importer", "exporter"}, edgeDir)(server)
		},
		func(ctx context.Context, mux *runtime.ServeMux, grpcEndpoint string, opts []grpc.DialOption) error {
			err := e.getAuthorizerGatewayRegistrations()(ctx, mux, grpcEndpoint, opts)
			if err != nil {
				return err
			}
			err = e.getEdgeGatewayRegistration([]string{"reader", "writer", "importer", "exporter"})(ctx, mux, grpcEndpoint, opts)
			if err != nil {
				return err
			}
			return nil
		},
		true)
	if err != nil {
		return nil, err
	}
	if server.Gateway.Mux != nil {
		// Add optional handlers.
		server.Gateway.Mux.Handle("/openapi.json", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			http.FileServer(http.FS(openapi.Static())).ServeHTTP(w, r)
		}))

		// attach handler for directory openapi spec when to same service as the authorizer.
		server.Gateway.Mux.Handle("/api/v2/directory/openapi.json", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			http.FileServer(http.FS(diropenapi.Static())).ServeHTTP(w, r)
		}))
	}
	return server, nil
}

func (e *Authorizer) createAuthorizer(cfg *builder.API, authorizerOpts []grpc.ServerOption) (*builder.Server, error) {
	server, err := e.ServiceBuilder.CreateService(cfg, authorizerOpts,
		e.getAuthorizerRegistration(),
		e.getAuthorizerGatewayRegistrations(), true)
	if err != nil {
		return nil, err
	}

	if server.Gateway.Mux != nil {
		// Add optional handlers.
		server.Gateway.Mux.Handle("/openapi.json", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			http.FileServer(http.FS(openapi.Static())).ServeHTTP(w, r)
		}))
	}
	return server, nil
}

func isLocalDirectory(address string) bool {
	return strings.Contains(address, "localhost") ||
		strings.Contains(address, "127.0.0.1") ||
		strings.Contains(address, "0.0.0.0")
}
