package app

import (
	"context"
	"strconv"
	"strings"

	edgeServer "github.com/aserto-dev/go-edge-ds/pkg/server"
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
	remoteConfig, err := e.Configuration.Directory.ToRemoteConfig()
	if err != nil {
		return err
	}
	edgeConfig, err := e.Configuration.Directory.ToEdgeConfig()
	if err != nil {
		return err
	}
	if (strings.Contains(remoteConfig.Address, "localhost") || strings.Contains(remoteConfig.Address, "0.0.0.0")) &&
		edgeConfig.DBPath != "" {
		addr := strings.Split(remoteConfig.Address, ":")
		if len(addr) != 2 {
			return errors.Errorf("invalid remote address - should contain <host>:<port>")
		}

		port, err := strconv.Atoi(addr[1])
		if err != nil {
			return err
		}

		edge, err := edgeServer.NewEdgeServer(
			*edgeConfig,
			&e.Configuration.API.GRPC.Certs,
			addr[0],
			port,
			e.Logger,
		)
		if err != nil {
			return errors.Wrap(err, "failed to create edge directory server")
		}

		e.Server.RegisterServer("edgeDirServer", edge.Start, edge.Stop)
	}

	err = e.Server.Start(e.Context)
	if err != nil {
		return errors.Wrap(err, "failed to start engine server")
	}

	return nil
}
