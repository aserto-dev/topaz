package app

import (
	"context"
	"strings"

	"github.com/aserto-dev/certs"
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
	if edgeConfig.Services == nil {
		edgeConfig.Services = make(map[string]*directory.API)
	}

	authorizerAPIConfig, ok := e.Configuration.Services["authorizer"]
	if !ok {
		return errors.New("invalid authorizer configuration")
	}

	if remoteConfig.Address == "" {
		e.Configuration.DirectoryResolver.Address = authorizerAPIConfig.GRPC.ListenAddress
		for key, config := range e.Configuration.Services {
			if key != "authorizer" {
				edgeConfig.Services[key] = config
			}
		}
		e.setEdgeDefaults(edgeConfig, authorizerAPIConfig.GRPC.Certs, "0.0.0.0:8282")
		if _, ok := edgeConfig.Services["reader"]; !ok {
			return errors.New("reader service must be configured")
		}
		if _, ok := edgeConfig.Services["reader"]; ok {
			if e.Configuration.DirectoryResolver.Address != edgeConfig.Services["reader"].GRPC.ListenAddress {
				return errors.New("remote address must match reader configuration address for the topaz directory resolver")
			}
		}
		edgeServer, err := edge.NewEdgeServer(*edgeConfig, &authorizerAPIConfig.GRPC.Certs, e.Logger)
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
		if len(edgeServer.Servers) > 0 {
			for name, edgeServer := range edgeServer.Servers {
				e.Server.RegisterServer(name, edgeServer.Start, edgeServer.Stop)
			}
		}
		return nil
	}

	if (strings.Contains(remoteConfig.Address, "localhost") || strings.Contains(remoteConfig.Address, "0.0.0.0")) && edgeConfig.DBPath != "" {
		for key, config := range e.Configuration.Services {
			if key != "authorizer" {
				edgeConfig.Services[key] = config
			}
		}
		e.setEdgeDefaults(edgeConfig, authorizerAPIConfig.GRPC.Certs, remoteConfig.Address)
		if _, ok := edgeConfig.Services["reader"]; !ok {
			return errors.New("reader service must be configured")
		}
		if _, ok := edgeConfig.Services["reader"]; ok {
			if e.Configuration.DirectoryResolver.Address != edgeConfig.Services["reader"].GRPC.ListenAddress {
				return errors.New("remote address must match reader configuration address for the topaz directory resolver")
			}
		}
		edgeServer, err := edge.NewEdgeServer(*edgeConfig, &authorizerAPIConfig.GRPC.Certs, e.Logger)
		if err != nil {
			return err
		}
		e.Server.RegisterServer("edgeServer", edgeServer.Start, edgeServer.Stop)
	}
	return nil
}

func (e *Authorizer) setEdgeDefaults(edgeConfig *directory.Config, defaultCerts certs.TLSCredsConfig, defaultRemoteAddress string) {
	if len(edgeConfig.Services) == 0 {
		// set defaults if not configred.
		defaultAPI := directory.API{}
		defaultAPI.GRPC.Certs = defaultCerts
		defaultAPI.GRPC.ListenAddress = defaultRemoteAddress // will only be used if remote address is specified.

		edgeConfig.Services = map[string]*directory.API{
			"reader":   &defaultAPI,
			"writer":   &defaultAPI,
			"exporter": &defaultAPI,
			"importer": &defaultAPI,
		}
	}
	// attach certs if not configured separately for the exposed services.
	for k, v := range edgeConfig.Services {
		if edgeConfig.Services[k].GRPC.Certs.TLSCACertPath == "" {
			v.GRPC.Certs = defaultCerts
			edgeConfig.Services[k] = v
		}
	}
}
