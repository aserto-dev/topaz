//go:build wireinject
// +build wireinject

package topaz

import (
	"github.com/google/wire"
	"google.golang.org/grpc"

	"github.com/aserto-dev/logger"
	"github.com/aserto-dev/topaz/pkg/app"
	"github.com/aserto-dev/topaz/pkg/cc"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	builder "github.com/aserto-dev/topaz/pkg/service/builder"
	"github.com/aserto-dev/topaz/resolvers"
)

var (
	commonSet = wire.NewSet(
		resolvers.New,

		builder.NewServiceFactory,
		builder.NewServiceManager,

		DefaultGRPCOptions,
		DefaultServices,

		wire.FieldsOf(new(*cc.CC), "Config", "Log", "Context", "ErrGroup"),
		wire.FieldsOf(new(*config.Config), "Common", "DecisionLogger"),
		wire.Struct(new(app.Topaz), "*"),
	)

	appTestSet = wire.NewSet(
		commonSet,
		cc.NewTestCC,
	)

	appSet = wire.NewSet(
		commonSet,
		cc.NewCC,
	)
)

func BuildApp(logOutput logger.Writer, errOutput logger.ErrWriter, configPath config.Path, overrides config.Overrider) (*app.Topaz, func(), error) {
	wire.Build(appSet)
	return &app.Topaz{}, func() {}, nil
}

func BuildTestApp(logOutput logger.Writer, errOutput logger.ErrWriter, configPath config.Path, overrides config.Overrider) (*app.Topaz, func(), error) {
	wire.Build(appTestSet)
	return &app.Topaz{}, func() {}, nil
}

func DefaultGRPCOptions() []grpc.ServerOption {
	return nil
}

func DefaultServices() map[string]app.ServiceTypes {
	return make(map[string]app.ServiceTypes)
}
