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
	if (strings.Contains(e.Configuration.Directory.Remote.Addr, "localhost") || strings.Contains(e.Configuration.Directory.Remote.Addr, "0.0.0.0")) &&
		e.Configuration.Directory.EdgeConfig.DBPath != "" {
		addr := strings.Split(e.Configuration.Directory.Remote.Addr, ":")
		if len(addr) != 2 {
			return errors.Errorf("invalid remote address - should contain <host>:<port>")
		}
		port, err := strconv.Atoi(addr[1])
		if err != nil {
			return err
		}
		edge := edgeServer.NewEdgeServer(
			e.Configuration.Directory.EdgeConfig,
			&e.Configuration.API.GRPC.Certs,
			addr[0],
			port,
			e.Logger,
		)

		e.Server.RegisterServer("edgeDirServer", edge.Start, edge.Stop)
	}
	err := e.Server.Start(e.Context)
	if err != nil {
		return errors.Wrap(err, "failed to start engine server")
	}

	return nil
}
