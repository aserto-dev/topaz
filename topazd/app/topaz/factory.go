package topaz

import (
	"github.com/aserto-dev/logger"
	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/topazd/app"
	"github.com/aserto-dev/topaz/topazd/cc"
	"github.com/aserto-dev/topaz/topazd/service/builder"
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
