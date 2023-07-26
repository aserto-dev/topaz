package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/aserto-dev/go-aserto/client"
	"github.com/aserto-dev/self-decision-logger/logger/self"
	decisionlog "github.com/aserto-dev/topaz/decision_log"
	"github.com/aserto-dev/topaz/decision_log/logger/file"
	"github.com/aserto-dev/topaz/decision_log/logger/nop"
	"github.com/aserto-dev/topaz/pkg/app/middlewares"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"

	edge "github.com/aserto-dev/go-edge-ds/pkg/server"

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
	Context        context.Context
	Logger         *zerolog.Logger
	ServerOptions  []grpc.ServerOption
	Configuration  *config.Config
	ServiceBuilder *builder.ServiceFactory
	Manager        *builder.ServiceManager
	Services       map[string]ServiceTypes
}

type ServiceTypes interface {
	RegisteredServices() []string
	GetServerOptions() []grpc.ServerOption
	GetGRPCRegistrations() builder.GRPCRegistrations
	GetGatewayRegistration() builder.HandlerRegistrations
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
	if readerConfig, ok := e.Configuration.Services["reader"]; ok {
		if readerConfig.GRPC.ListenAddress != e.Configuration.DirectoryResolver.Address {
			return errors.New("remote directory resolver address is different from reader grpc address")
		}
	}

	dir, err := locker.New(&e.Configuration.Edge, e.Logger)
	if err != nil {
		return err
	}

	e.Services = make(map[string]ServiceTypes)

	serviceMap := mapToGRPCPorts(e.Configuration.Services)

	for address, config := range serviceMap {
		serviceConfig := config

		// get middlewares for edge services.
		middlewareList, err := middlewares.GetMiddlewaresForService(config.registeredServices[0], e.Context, e.Configuration, e.Logger)
		if err != nil {
			return err
		}

		var opts []grpc.ServerOption
		unary, streeam := middlewareList.AsGRPCOptions()
		opts = append(opts, unary)
		opts = append(opts, streeam)

		edge, err := NewEdgeDir(serviceConfig.registeredServices, serviceConfig.API, opts, dir)
		if err != nil {
			return err
		}
		e.Services["edge_"+address] = edge
		if contains(serviceConfig.registeredServices, "authorizer") {
			topaz, err := NewTopaz(serviceConfig.API, &e.Configuration.Common, opts, e.Logger)
			if err != nil {
				return err
			}
			e.Services["topaz"] = topaz
		}

		var grpcs []builder.GRPCRegistrations
		var gateways []builder.HandlerRegistrations

		for _, serv := range e.Services {
			notAdded := true
			for _, serviceName := range serv.RegisteredServices() {
				if contains(serviceConfig.registeredServices, serviceName) && notAdded {
					opts = append(opts, serv.GetServerOptions()...)
					grpcs = append(grpcs, serv.GetGRPCRegistrations())
					gateways = append(gateways, serv.GetGatewayRegistration())
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

func isLocalDirectory(address string) bool {
	return strings.Contains(address, "localhost") ||
		strings.Contains(address, "127.0.0.1") ||
		strings.Contains(address, "0.0.0.0")
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
