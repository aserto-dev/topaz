package app

import (
	"context"
	"net/http"

	console "github.com/aserto-dev/go-topaz-ui"
	cerr "github.com/aserto-dev/topaz/errors"
	"github.com/aserto-dev/topaz/internal/eds"
	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/topazd/app/handlers"
	"github.com/aserto-dev/topaz/topazd/app/middlewares"
	"github.com/aserto-dev/topaz/topazd/authentication"
	"github.com/aserto-dev/topaz/topazd/service/builder"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

// Topaz is an authorizer service instance, responsible for managing
// the authorizer API, user directory instance and the OPA plugins.
type Topaz struct {
	Context        context.Context
	Logger         *zerolog.Logger
	ServerOptions  []grpc.ServerOption
	Configuration  *config.Config
	ServiceBuilder *builder.ServiceFactory
	Manager        *builder.ServiceManager
	Services       map[string]builder.ServiceTypes
}

var healthCheck *health.Server

func SetServiceStatus(log *zerolog.Logger, service string, servingStatus grpc_health_v1.HealthCheckResponse_ServingStatus) {
	if healthCheck == nil {
		return
	}

	resp, err := healthCheck.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{Service: service})
	if err != nil {
		log.Error().Err(err).Str("service", service).Str("status", servingStatus.String()).Msg("health")
		return
	}

	// only write log message when the health state changed.
	if resp.GetStatus() != servingStatus {
		log.Info().Str("service", service).Str("status", servingStatus.String()).Msg("health")
		healthCheck.SetServingStatus(service, servingStatus)
	}
}

// Start starts all services required by the engine.
func (e *Topaz) Start() error {
	// build dependencies map.
	for _, cfg := range e.Configuration.APIConfig.Services {
		if len(cfg.Needs) > 0 {
			for _, name := range cfg.Needs {
				if dependencyConfig, ok := e.Configuration.APIConfig.Services[name]; ok {
					if !lo.Contains(e.Manager.DependencyMap[cfg.GRPC.ListenAddress], dependencyConfig.GRPC.ListenAddress) &&
						cfg.GRPC.ListenAddress != dependencyConfig.GRPC.ListenAddress {
						e.Manager.DependencyMap[cfg.GRPC.ListenAddress] = append(e.Manager.DependencyMap[cfg.GRPC.ListenAddress], dependencyConfig.GRPC.ListenAddress)
					}
				}
			}
		}
	}

	err := e.Manager.StartServers(e.Context)
	if err != nil {
		return errors.Wrap(err, "failed to start engine server")
	}

	// Add registered services to the health service
	if e.Manager.HealthServer != nil {
		healthCheck = e.Manager.HealthServer.Server

		for serviceName := range e.Configuration.APIConfig.Services {
			e.Manager.HealthServer.SetServiceStatus(serviceName, grpc_health_v1.HealthCheckResponse_SERVING)
		}

		// register phony sync service with status NOT_SERVING
		service, servingStatus := "sync", grpc_health_v1.HealthCheckResponse_NOT_SERVING
		e.Manager.HealthServer.Server.SetServingStatus(service, servingStatus)
		e.Logger.Info().Str("component", "edge.plugin").Str("service", service).Str("status", servingStatus.String()).Msg("health")
	}

	return nil
}

//nolint:funlen,nestif,gocognit
func (e *Topaz) ConfigServices() error {
	metricsMiddleware, err := e.setupHealthAndMetrics()
	if err != nil {
		return err
	}

	if err := e.prepareServices(); err != nil {
		return err
	}

	if err := e.validateConfig(); err != nil {
		return err
	}

	serviceMap := mapToGRPCPorts(e.Configuration.APIConfig.Services)

	for address, cfg := range serviceMap {
		e.Logger.Debug().Msgf("configuring address %s", address)

		serviceConfig := cfg

		// get middlewares for edge services.
		opts, err := middlewares.GetMiddlewaresForService(e.Context, e.Configuration, e.Logger)
		if err != nil {
			return err
		}

		opts = append(opts, metricsMiddleware...)

		var (
			grpcs    []builder.GRPCRegistrations
			gateways []builder.HandlerRegistrations
			cleanup  func()
		)

		for _, serv := range e.Services {
			added := false

			for _, serviceName := range serv.AvailableServices() {
				if added || !lo.Contains(serviceConfig.registeredServices, serviceName) {
					continue
				}

				grpcs = append(grpcs, serv.GetGRPCRegistrations(serviceConfig.registeredServices...))

				gatewayPort, err := config.PortFromAddress(cfg.API.Gateway.ListenAddress)
				if err != nil {
					return errors.Wrapf(err, "invalid gateway address %q in service %q", cfg.API.Gateway.ListenAddress, serviceName)
				}

				gateways = append(gateways, serv.GetGatewayRegistration(gatewayPort, serviceConfig.registeredServices...))
				cleanup = serv.Close
				added = true
			}
		}

		server, err := e.ServiceBuilder.CreateService(
			serviceConfig.API,
			&builder.GRPCOptions{
				ServerOptions: opts,
				Registrations: func(server *grpc.Server) {
					for _, f := range grpcs {
						f(server)
					}
				},
			},
			&builder.GatewayOptions{
				HandlerRegistrations: func(ctx context.Context, mux *runtime.ServeMux, grpcEndpoint string, opts []grpc.DialOption) error {
					for _, f := range gateways {
						if err := f(ctx, mux, grpcEndpoint, opts); err != nil {
							return err
						}
					}

					return nil
				},
				ErrorHandler: cerr.CustomErrorHandler,
			},
			cleanup,
		)
		if err != nil {
			return err
		}

		if readerSvc, ok := e.Configuration.APIConfig.Services["reader"]; ok {
			if err := handlers.SetWellKnownConfigHandler(readerSvc, server.Gateway.Mux); err != nil {
				return err
			}
		}

		if con, ok := e.Services[consoleService]; ok {
			if lo.Contains(serviceConfig.registeredServices, consoleService) {
				if server.Gateway != nil && server.Gateway.Mux != nil {
					apiKeyAuthMiddleware, err := authentication.NewAPIKeyAuthMiddleware(e.Context, &e.Configuration.Auth, e.Logger)
					if err != nil {
						return err
					}

					consoleSvc, ok := con.(*ConsoleService)
					if !ok {
						return status.Errorf(codes.Internal, "failed type assertion %q", "ConsoleService")
					}

					consoleConfig := consoleSvc.PrepareConfig(e.Configuration)

					// config service.
					server.Gateway.Mux.Handle("/api/v2/config", apiKeyAuthMiddleware.ConfigAuth(handlers.ConfigHandlerV2(consoleConfig), e.Configuration.Auth))
					server.Gateway.Mux.HandleFunc("/api/v1/authorizers", handlers.AuthorizersHandler(consoleConfig))
					// console service. depends on config service.
					server.Gateway.Mux.Handle("/ui/", handlers.UIHandler(http.FS(console.FS)))
					server.Gateway.Mux.Handle("/assets/", handlers.UIHandler(http.FS(console.FS)))
				}
			}
		}

		if err := e.Manager.AddGRPCServer(server); err != nil {
			return err
		}
	}

	return nil
}

func (e *Topaz) setupHealthAndMetrics() ([]grpc.ServerOption, error) {
	if e.Configuration.APIConfig.Health.ListenAddress != "" {
		err := e.Manager.SetupHealthServer(e.Configuration.APIConfig.Health.ListenAddress, &e.Configuration.APIConfig.Health.Certificates)
		if err != nil {
			return nil, err
		}
	}

	if e.Configuration.APIConfig.Metrics.ListenAddress != "" {
		metricsMiddleware, err := e.Manager.SetupMetricsServer(e.Configuration.APIConfig.Metrics.ListenAddress,
			&e.Configuration.APIConfig.Metrics.Certificates,
			false)
		if err != nil {
			return nil, err
		}

		return metricsMiddleware, nil
	}

	return nil, nil
}

func (e *Topaz) prepareServices() error {
	// prepare services
	if e.Configuration.Edge.DBPath != "" {
		dir, err := eds.New(e.Context, &e.Configuration.Edge, e.Logger)
		if err != nil {
			return err
		}

		edgeDir, err := NewEdgeDir(dir)
		if err != nil {
			return err
		}

		e.Services["edge"] = edgeDir
	}

	if serviceConfig, ok := e.Configuration.APIConfig.Services[authorizerService]; ok {
		authorizer, err := NewAuthorizer(e.Context, serviceConfig, &e.Configuration.Common, nil, e.Logger)
		if err != nil {
			return err
		}

		e.Services["authorizer"] = authorizer
	}

	if _, ok := e.Configuration.APIConfig.Services[consoleService]; ok {
		e.Services["console"] = NewConsole()
	}

	return nil
}

type services struct {
	registeredServices []string
	API                *builder.API
}

func mapToGRPCPorts(api map[string]*builder.API) map[string]services {
	portMap := make(map[string]services)

	for key, config := range api {
		if existing, ok := portMap[config.GRPC.ListenAddress]; ok {
			existing.registeredServices = append(existing.registeredServices, key)
			if existing.API.Gateway.ListenAddress == "" && config.Gateway.ListenAddress != "" {
				existing.API.Gateway = config.Gateway
			}

			portMap[config.GRPC.ListenAddress] = existing
		} else {
			serv := services{}
			serv.registeredServices = append(serv.registeredServices, key)
			serv.API = config
			portMap[config.GRPC.ListenAddress] = serv
		}
	}

	return portMap
}

func (e *Topaz) validateConfig() error {
	if readerConfig, ok := e.Configuration.APIConfig.Services["reader"]; ok {
		if readerConfig.GRPC.ListenAddress != e.Configuration.DirectoryResolver.Address {
			for _, serviceName := range e.Services["edge"].AvailableServices() {
				delete(e.Configuration.APIConfig.Services, serviceName)
			}

			delete(e.Services, "edge")

			e.Logger.Info().Msg("disabling local directory services")
		}
	}

	if _, ok := e.Configuration.APIConfig.Services["console"]; ok {
		if _, ok := e.Configuration.APIConfig.Services["reader"]; !ok {
			return errors.New("console needs the reader service to be configured")
		}
	}

	for key := range e.Configuration.APIConfig.Services {
		validKey := false

		for _, service := range e.Services {
			if lo.Contains(service.AvailableServices(), key) {
				validKey = true
				break
			}
		}

		if !validKey {
			return errors.Errorf("unknown service type %s", key)
		}
	}

	return nil
}
