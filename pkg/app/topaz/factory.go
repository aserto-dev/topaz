package topaz

import (
	"github.com/aserto-dev/logger"
	"github.com/aserto-dev/topaz/pkg/app"
	"github.com/aserto-dev/topaz/pkg/cc"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/service/builder"
	"google.golang.org/grpc"
)

func BuildApp(
	logOutput logger.Writer,
	errOutput logger.ErrWriter,
	configPath config.Path,
	overrides config.Overrider,
) (
	*app.Topaz,
	func(),
	error,
) {
	ccCC, cleanup, err := cc.NewCC(logOutput, errOutput, configPath, overrides)
	if err != nil {
		return nil, nil, err
	}

	context := ccCC.Context
	zerologLogger := ccCC.Log
	v := DefaultGRPCOptions()
	configConfig := ccCC.Config
	serviceFactory := builder.NewServiceFactory()
	serviceManager := builder.NewServiceManager(zerologLogger)
	v2 := DefaultServices()
	topaz := &app.Topaz{
		Context:        context,
		Logger:         zerologLogger,
		ServerOptions:  v,
		Configuration:  configConfig,
		ServiceBuilder: serviceFactory,
		Manager:        serviceManager,
		Services:       v2,
	}

	return topaz, func() {
		cleanup()
	}, nil
}

func DefaultGRPCOptions() []grpc.ServerOption {
	return nil
}

func DefaultServices() map[string]builder.ServiceTypes {
	return make(map[string]builder.ServiceTypes)
}
