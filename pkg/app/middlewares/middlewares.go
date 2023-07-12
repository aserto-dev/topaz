package middlewares

import (
	"github.com/aserto-dev/aserto-grpc/grpcutil"
	metricserver "github.com/aserto-dev/aserto-grpc/grpcutil/metrics"
	"github.com/aserto-dev/aserto-grpc/grpcutil/middlewares/gerr"
	"github.com/aserto-dev/aserto-grpc/grpcutil/middlewares/metrics"
	"github.com/aserto-dev/aserto-grpc/grpcutil/middlewares/request"
	"github.com/aserto-dev/aserto-grpc/grpcutil/middlewares/tracing"
	"github.com/aserto-dev/topaz/pkg/app"
	"github.com/aserto-dev/topaz/pkg/app/auth"
	"github.com/rs/zerolog"
)

func AttachMiddlewares(app *app.Authorizer) error {

	var middlewareList []grpcutil.Middleware
	if len(app.Configuration.Auth.APIKeys) > 0 {
		authmiddleware, err := auth.NewAPIKeyAuthMiddleware(app.Context, &app.Configuration.Auth, app.Logger)
		if err != nil {
			return err
		}

		middlewareList = append(middlewareList, authmiddleware)
	}

	middlewareList = append(middlewareList, request.NewRequestIDMiddleware())
	middlewareList = append(middlewareList, tracing.NewTracingMiddleware(app.Logger))
	middlewareList = append(middlewareList, gerr.NewErrorMiddleware())

	middlewares := metrics.NewMiddlewares(app.Configuration.Metrics,
		middlewareList...)

	// start metrics http server if configured.
	if app.Configuration.Metrics.ListenAddress != "" {
		go startMetricsHTTPServer(&metricserver.Config{
			ListenAddress: app.Configuration.Metrics.ListenAddress,
			ZPages:        app.Configuration.Metrics.ZPages,
			GRPC:          app.Configuration.Metrics.GRPC,
			HTTP:          app.Configuration.Metrics.HTTP,
			DB:            app.Configuration.Metrics.DB,
		}, app.Logger)
	}

	app.AddGRPCServerOptions(middlewares.AsGRPCOptions())

	return nil
}

func startMetricsHTTPServer(config *metricserver.Config, logger *zerolog.Logger) {
	msrv := metricserver.NewServer(config, logger)
	err := msrv.HTTP().ListenAndServe()
	if err != nil {
		logger.Error().Err(err).Msg("failed to start metrics server")
	}
}
