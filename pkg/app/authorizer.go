package app

import (
	"context"
	"strings"

	"github.com/aserto-dev/go-edge-ds/pkg/directory"
	edge "github.com/aserto-dev/go-edge-ds/pkg/server"
	"github.com/aserto-dev/topaz/pkg/app/server"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
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
	remoteConfig, err := e.Configuration.Directory.ToRemoteConfig()
	if err != nil {
		return err
	}
	edgeConfig, err := e.Configuration.Directory.ToEdgeConfig()
	if err != nil {
		return err
	}

	if _, ok := edgeConfig.Services["reader"]; ok {
		if remoteConfig.Address != edgeConfig.Services["reader"].GRPC.ListenAddress {
			return errors.New("remote address must match reader configuration address for the topaz directory resolver")
		}
	}
	if (strings.Contains(remoteConfig.Address, "localhost") || strings.Contains(remoteConfig.Address, "0.0.0.0")) && edgeConfig.DBPath != "" {

		if len(edgeConfig.Services) == 0 {
			//set defaults
			defaultAPI := directory.API{}
			defaultAPI.GRPC.Certs = e.Configuration.API.GRPC.Certs
			defaultAPI.GRPC.ListenAddress = "localhost:9292"
			defaultAPI.GRPC.ConnectionTimeoutSeconds = e.Configuration.API.GRPC.ConnectionTimeoutSeconds
			defaultAPI.Gateway.AllowedOrigins = e.Configuration.API.Gateway.AllowedOrigins
			defaultAPI.Gateway.Certs = e.Configuration.API.Gateway.Certs
			defaultAPI.Gateway.ListenAddress = "localhost:9293"

			edgeConfig.Services = map[string]*directory.API{
				"reader":   &defaultAPI,
				"writer":   &defaultAPI,
				"exporter": &defaultAPI,
				"importer": &defaultAPI,
			}
		}
		//attach certs if not configured separately for the exposed services
		for k, v := range edgeConfig.Services {
			if edgeConfig.Services[k].GRPC.Certs.TLSCACertPath == "" {
				v.GRPC.Certs = e.Configuration.API.GRPC.Certs
				edgeConfig.Services[k] = v
			}
		}
		edgeServer, err := edge.NewEdgeServer(*edgeConfig, &e.Configuration.API.GRPC.Certs, e.Logger)
		if err != nil {
			return err
		}
		e.Server.RegisterServer("edgeServer", edgeServer.Start, edgeServer.Stop)
	}
	return nil
}
