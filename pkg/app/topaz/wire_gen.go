// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package topaz

import (
	"github.com/aserto-dev/logger"
	"github.com/aserto-dev/topaz/pkg/app"
	"github.com/aserto-dev/topaz/pkg/app/auth"
	"github.com/aserto-dev/topaz/pkg/app/impl"
	"github.com/aserto-dev/topaz/pkg/builder"
	"github.com/aserto-dev/topaz/pkg/cc"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/google/wire"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

// Injectors from wire.go:

func BuildApp(logOutput logger.Writer, errOutput logger.ErrWriter, configPath config.Path, overrides config.Overrider) (*app.Authorizer, func(), error) {
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
	resolversResolvers := resolvers.New()
	common := &configConfig.Common
	authorizerServer := impl.NewAuthorizerServer(zerologLogger, common, resolversResolvers)
	authorizer := &app.Authorizer{
		Context:          context,
		Logger:           zerologLogger,
		ServerOptions:    v,
		Configuration:    configConfig,
		ServiceBuilder:   serviceFactory,
		Manager:          serviceManager,
		Resolver:         resolversResolvers,
		AuthorizerServer: authorizerServer,
	}
	return authorizer, func() {
		cleanup()
	}, nil
}

func BuildTestApp(logOutput logger.Writer, errOutput logger.ErrWriter, configPath config.Path, overrides config.Overrider) (*app.Authorizer, func(), error) {
	ccCC, cleanup, err := cc.NewTestCC(logOutput, errOutput, configPath, overrides)
	if err != nil {
		return nil, nil, err
	}
	context := ccCC.Context
	zerologLogger := ccCC.Log
	v := DefaultGRPCOptions()
	configConfig := ccCC.Config
	serviceFactory := builder.NewServiceFactory()
	serviceManager := builder.NewServiceManager(zerologLogger)
	resolversResolvers := resolvers.New()
	common := &configConfig.Common
	authorizerServer := impl.NewAuthorizerServer(zerologLogger, common, resolversResolvers)
	authorizer := &app.Authorizer{
		Context:          context,
		Logger:           zerologLogger,
		ServerOptions:    v,
		Configuration:    configConfig,
		ServiceBuilder:   serviceFactory,
		Manager:          serviceManager,
		Resolver:         resolversResolvers,
		AuthorizerServer: authorizerServer,
	}
	return authorizer, func() {
		cleanup()
	}, nil
}

// wire.go:

var (
	commonSet = wire.NewSet(resolvers.New, impl.NewAuthorizerServer, auth.NewAPIKeyAuthMiddleware, builder.NewServiceFactory, builder.NewServiceManager, DefaultGRPCOptions, wire.FieldsOf(new(*cc.CC), "Config", "Log", "Context", "ErrGroup"), wire.FieldsOf(new(*config.Config), "Common", "DecisionLogger"), wire.Struct(new(app.Authorizer), "*"))

	appTestSet = wire.NewSet(
		commonSet, cc.NewTestCC, prometheus.NewRegistry, wire.Bind(new(prometheus.Registerer), new(*prometheus.Registry)),
	)

	appSet = wire.NewSet(
		commonSet, cc.NewCC, wire.InterfaceValue(new(prometheus.Registerer), prometheus.DefaultRegisterer),
	)
)

func DefaultGRPCOptions() []grpc.ServerOption {
	return nil
}
