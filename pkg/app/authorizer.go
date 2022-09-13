package app

import (
	"context"

	"github.com/aserto-dev/topaz/pkg/app/server"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Authorizer is an Aserto Edge Authorizer instance
// It's responsible with managing the Aserto Edge API, the User Directory
// and the OPA plugins
type Authorizer struct {
	Context         context.Context
	Logger          *zerolog.Logger
	Configuration   *config.Common
	Server          *server.Server
	RuntimeResolver resolvers.RuntimeResolver
}

// Start starts all services required by the Engine
func (e *Authorizer) Start() error {
	err := e.Server.Start(e.Context)
	if err != nil {
		return errors.Wrap(err, "failed to start engine server")
	}

	return nil
}
