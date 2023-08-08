package app

import (
	"context"
	"fmt"
	"time"

	"github.com/aserto-dev/go-aserto/client"
	"github.com/aserto-dev/self-decision-logger/logger/self"
	decisionlog "github.com/aserto-dev/topaz/decision_log"
	"github.com/aserto-dev/topaz/decision_log/logger/file"
	"github.com/aserto-dev/topaz/decision_log/logger/nop"
	"github.com/aserto-dev/topaz/pkg/app/middlewares"
	"github.com/aserto-dev/topaz/pkg/cc/config"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"

	edge "github.com/aserto-dev/go-edge-ds/pkg/server"

	builder "github.com/aserto-dev/service-host"
)

var locker edge.EdgeDirLock

// Authorizer is an authorizer service instance, responsible for managing
// the authorizer API, user directory instance and the OPA plugins.
type Authorizer struct {
	Context        context.Context
	Logger         *zerolog.Logger
	ServerOptions  []grpc.ServerOption
	Configuration  *config.Config
	ServiceBuilder *builder.ServiceFactory
	Manager        *builder.ServiceManager
	Services       map[string]ServiceTypes
}

type ServiceTypes interface {
	AvailableServices() []string
	GetGRPCRegistrations(services ...string) builder.GRPCRegistrations
	GetGatewayRegistration(services ...string) builder.HandlerRegistrations
}

func (e *Authorizer) AddGRPCServerOptions(grpcOptions ...grpc.ServerOption) {
	e.ServerOptions = append(e.ServerOptions, grpcOptions...)
}

// Start starts all services required by the engine.
func (e *Authorizer) Start() error {
	// build dependencies map.
	for _, cfg := range e.Configuration.Services {
		if len(cfg.Needs) > 0 {
			for _, name := range cfg.Needs {
				if dependencyConfig, ok := e.Configuration.Services[name]; ok {
					if !contains(e.Manager.DependencyMap[cfg.GRPC.ListenAddress], dependencyConfig.GRPC.ListenAddress) &&
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

	return nil
}

func (e *Authorizer) ConfigServices() error {

	// prepare services
	dir, err := locker.New(&e.Configuration.Edge, e.Logger)
	if err != nil {
		return err
	}

	edgeDir, err := NewEdgeDir(dir)
	if err != nil {
		return err
	}
	e.Services["edge"] = edgeDir

	if serviceConfig, ok := e.Configuration.Services[authorizerService]; ok {
		topaz, err := NewTopaz(serviceConfig, &e.Configuration.Common, nil, e.Logger)
		if err != nil {
			return err
		}
		e.Services["topaz"] = topaz
	}

	if err := e.validateConfig(); err != nil {
		return err
	}

	serviceMap := mapToGRPCPorts(e.Configuration.Services)

	for address, config := range serviceMap {
		e.Logger.Debug().Msgf("configuring address %s", address)
		serviceConfig := config

		// get middlewares for edge services.
		opts, err := middlewares.GetMiddlewaresForService(e.Context, e.Configuration, e.Logger)
		if err != nil {
			return err
		}

		var grpcs []builder.GRPCRegistrations
		var gateways []builder.HandlerRegistrations

		for _, serv := range e.Services {
			notAdded := true
			for _, serviceName := range serv.AvailableServices() {
				if contains(serviceConfig.registeredServices, serviceName) && notAdded {
					grpcs = append(grpcs, serv.GetGRPCRegistrations(serviceConfig.registeredServices...))
					gateways = append(gateways, serv.GetGatewayRegistration(serviceConfig.registeredServices...))
					notAdded = false
				}
			}
		}
		server, err := e.ServiceBuilder.CreateService(serviceConfig.API,
			opts,
			func(server *grpc.Server) {
				for _, f := range grpcs {
					f(server)
				}
			},
			func(ctx context.Context, mux *runtime.ServeMux, grpcEndpoint string, opts []grpc.DialOption) error {
				for _, f := range gateways {
					err := f(ctx, mux, grpcEndpoint, opts)
					if err != nil {
						return err
					}
				}
				return nil
			}, true)

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
		// set default gateway timeouts
		if serv.API.Gateway.ReadTimeout == 0 {
			serv.API.Gateway.ReadTimeout = 2 * time.Second
		}
		if serv.API.Gateway.ReadHeaderTimeout == 0 {
			serv.API.Gateway.ReadHeaderTimeout = 2 * time.Second
		}
		if serv.API.Gateway.WriteTimeout == 0 {
			serv.API.Gateway.WriteTimeout = 2 * time.Second
		}
		if serv.API.Gateway.IdleTimeout == 0 {
			serv.API.Gateway.WriteTimeout = 30 * time.Second
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

func (e *Authorizer) GetDecisionLogger(cfg config.DecisionLogConfig) (decisionlog.DecisionLogger, error) {
	var decisionlogger decisionlog.DecisionLogger
	var err error

	switch cfg.Type {
	case "self":
		decisionlogger, err = self.New(e.Context, cfg.Config, e.Logger, client.NewDialOptionsProvider())
		if err != nil {
			return nil, err
		}

	case "file":
		maxsize := 0
		maxfiles := 0

		logpath := cfg.Config["log_file_path"]
		maxsize, _ = cfg.Config["max_file_size_mb"].(int)
		maxfiles, _ = cfg.Config["max_file_count"].(int)

		decisionlogger, err = file.New(e.Context, &file.Config{
			LogFilePath:   fmt.Sprintf("%v", logpath),
			MaxFileSizeMB: maxsize,
			MaxFileCount:  maxfiles,
		}, e.Logger)
		if err != nil {
			return nil, err
		}

	default:
		decisionlogger, err = nop.New(e.Context, e.Logger)
		if err != nil {
			return nil, err
		}

	}

	return decisionlogger, err
}

func (e *Authorizer) validateConfig() error {
	if readerConfig, ok := e.Configuration.Services["reader"]; ok {
		if readerConfig.GRPC.ListenAddress != e.Configuration.DirectoryResolver.Address {
			return errors.New("remote directory resolver address is different from reader grpc address")
		}
	}

	for key := range e.Configuration.Services {
		if !(contains(e.Services["edge"].AvailableServices(), key) || key == authorizerService) {
			return errors.Errorf("unknown service type %s", key)
		}
	}
	return nil
}
