package app

import (
	"context"
	"strings"

	"github.com/aserto-dev/go-edge-ds/pkg/directory"
	edge "github.com/aserto-dev/go-edge-ds/pkg/server"
	"github.com/aserto-dev/topaz/pkg/app/server"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

// Authorizer is an authorizer service instance, responsible for managing
// the authorizer API, user directory instance and the OPA plugins.
type Authorizer struct {
	Context       context.Context
	Logger        *zerolog.Logger
	Configuration *config.Config
	Server        *server.Server
	Resolver      *resolvers.Resolvers
}

// Start starts all services required by the engine.
func (e *Authorizer) Start() error {
	err := e.configEdge()
	if err != nil {
		return err
	}

	err = e.Server.Start(e.Context)
	if err != nil {
		return errors.Wrap(err, "failed to start engine server")
	}

	return nil
}

func (e *Authorizer) configEdge() error {
	// Directory client configuration.
	remoteConfig := e.Configuration.DirectoryResolver

	// Edge configuration
	edgeConfig := &e.Configuration.Edge

	authorizerAPIConfig, ok := e.Configuration.Services["authorizer"]
	if !ok {
		return errors.New("invalid authorizer configuration")
	}

	if remoteConfig.Address == "" {
		e.Configuration.DirectoryResolver.Address = authorizerAPIConfig.GRPC.ListenAddress

		edgeServer, err := edge.NewEdgeServer([]string{"reader", "writer", "importer", "exporter"}, *edgeConfig, authorizerAPIConfig, e.Logger)
		if err != nil {
			return err
		}
		registeredGRPC := e.Server.GRPCRegistrations
		e.Server.GRPCRegistrations = func(server *grpc.Server) {
			edgeServer.GetGRPCRegistrations()(server)
			registeredGRPC(server)
		}
		registeredGateway := e.Server.HandlerRegistrations
		e.Server.HandlerRegistrations = func(ctx context.Context, mux *runtime.ServeMux, grpcEndpoint string, opts []grpc.DialOption) error {
			err := registeredGateway(ctx, mux, grpcEndpoint, opts)
			if err != nil {
				return err
			}
			err = edgeServer.GetGatewayRegistrations()(ctx, mux, grpcEndpoint, opts)
			if err != nil {
				return err
			}
			return nil
		}
		return nil
	}

	if (strings.Contains(remoteConfig.Address, "localhost") || strings.Contains(remoteConfig.Address, "0.0.0.0")) && edgeConfig.DBPath != "" {
		if readerConfig, ok := e.Configuration.Services["reader"]; ok {
			if readerConfig.GRPC.ListenAddress != remoteConfig.Address {
				return errors.New("directory resolver address must match the reader grpc address")
			}
		} else {
			return errors.New("reader not configured")
		}

		if len(e.Configuration.Services) > 1 {
			serviceMap, err := mapToGRPCPorts(e.Configuration.Services)
			if err != nil {
				return err
			}
			for key, config := range serviceMap {
				if config.registerName[0] != "authorizer" {
					if config.API.GRPC.Certs.TLSCACertPath == "" {
						config.API.GRPC.Certs = authorizerAPIConfig.GRPC.Certs
					}
					if config.API.Gateway.Certs.TLSCACertPath == "" {
						config.API.Gateway.HTTP = true
					}
					edgeServer, err := edge.NewEdgeServer(config.registerName, *edgeConfig, config.API, e.Logger)
					if err != nil {
						return err
					}
					e.Server.RegisterServer(key, edgeServer.Start, edgeServer.Stop)
				}
			}
		} else {
			defaultAPI := directory.API{}
			defaultAPI.GRPC.ListenAddress = remoteConfig.Address
			defaultAPI.GRPC.Certs = authorizerAPIConfig.GRPC.Certs // use same certs as authorizer.
			defaultAPI.Gateway.ListenAddress = ""                  // no gateway initialized by default.
			edgeServer, err := edge.NewEdgeServer([]string{"reader", "writer", "importer", "exporter"}, *edgeConfig, &defaultAPI, e.Logger)
			if err != nil {
				return err
			}
			e.Server.RegisterServer("edge-directory", edgeServer.Start, edgeServer.Stop)
		}
	}
	return nil
}

type services struct {
	registerName []string
	API          *directory.API
}

func mapToGRPCPorts(api map[string]*directory.API) (map[string]services, error) {
	portMap := make(map[string]services)
	for key, config := range api {
		serv := services{}
		if existing, ok := portMap[config.GRPC.ListenAddress]; ok {
			serv = existing
			serv.registerName = append(serv.registerName, key)
		} else {
			serv.registerName = append(serv.registerName, key)
			serv.API = config
		}
		portMap[config.GRPC.ListenAddress] = serv
	}
	return portMap, nil
}
