//go:build wireinject
// +build wireinject

package topaz

import (
	"github.com/google/wire"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/aserto-dev/logger"
	"github.com/aserto-dev/topaz/pkg/app"
	"github.com/aserto-dev/topaz/pkg/app/auth"
	"github.com/aserto-dev/topaz/pkg/app/impl"
	"github.com/aserto-dev/topaz/pkg/app/server"
	"github.com/aserto-dev/topaz/pkg/cc"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/resolvers"
)

var (
	commonSet = wire.NewSet(
		server.NewServer,
		server.NewGatewayServer,
		server.GatewayMux,

		resolvers.New,
		impl.NewAuthorizerServer,

		GRPCServerRegistrations,
		GatewayServerRegistrations,

		auth.NewAPIKeyAuthMiddleware,

		wire.FieldsOf(new(*cc.CC), "Config", "Log", "Context", "ErrGroup"),
		wire.FieldsOf(new(*config.Config), "Common", "DecisionLogger"),
		wire.Struct(new(app.Authorizer), "*"),
	)

	appTestSet = wire.NewSet(
		commonSet,
		cc.NewTestCC,
		prometheus.NewRegistry,
		wire.Bind(new(prometheus.Registerer), new(*prometheus.Registry)),
	)

	appSet = wire.NewSet(
		commonSet,
		cc.NewCC,
		wire.InterfaceValue(new(prometheus.Registerer), prometheus.DefaultRegisterer),
	)
)

func BuildApp(logOutput logger.Writer, errOutput logger.ErrWriter, configPath config.Path, overrides config.Overrider) (*app.Authorizer, func(), error) {
	wire.Build(appSet)
	return &app.Authorizer{}, func() {}, nil
}

func BuildTestApp(logOutput logger.Writer, errOutput logger.ErrWriter, configPath config.Path, overrides config.Overrider) (*app.Authorizer, func(), error) {
	wire.Build(appTestSet)
	return &app.Authorizer{}, func() {}, nil
}
