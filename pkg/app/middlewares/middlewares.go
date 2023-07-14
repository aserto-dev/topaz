package middlewares

import (
	"context"
	"fmt"

	"github.com/aserto-dev/aserto-grpc/grpcutil"
	"github.com/aserto-dev/aserto-grpc/grpcutil/middlewares/gerr"
	"github.com/aserto-dev/aserto-grpc/grpcutil/middlewares/metrics"
	"github.com/aserto-dev/aserto-grpc/grpcutil/middlewares/request"
	"github.com/aserto-dev/aserto-grpc/grpcutil/middlewares/tracing"
	"github.com/aserto-dev/go-edge-ds/pkg/session"
	"github.com/aserto-dev/topaz/pkg/app/auth"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/rs/zerolog"
)

func GetMiddlewaresForService(serviceName string, ctx context.Context, cfg *config.Config, logger *zerolog.Logger) (grpcutil.Middlewares, error) {

	if _, ok := cfg.Services[serviceName]; !ok {
		return nil, fmt.Errorf("Service %s not configured", serviceName)
	}
	var middlewareList grpcutil.Middlewares
	if len(cfg.Auth.APIKeys) > 0 {
		authmiddleware, err := auth.NewAPIKeyAuthMiddleware(ctx, &cfg.Auth, logger)
		if err != nil {
			return nil, err
		}

		middlewareList = append(middlewareList, authmiddleware)
	}
	if serviceName != "authorizer" {
		sessionMiddleware := session.HeaderMiddleware{DisableValidation: false}
		middlewareList = append(middlewareList, &sessionMiddleware)
	}
	middlewareList = append(middlewareList, request.NewRequestIDMiddleware())
	// only attach policy instance information if discovery resource is configured.
	if cfg.OPA.Config.Discovery != nil && cfg.OPA.Config.Discovery.Resource != nil {
		middlewareList = append(middlewareList, NewInstanceMiddleware(cfg, logger))
	}
	// get tenant id from opa instance id.
	middlewareList = append(middlewareList, NewTenantIDMiddleware(cfg))
	middlewareList = append(middlewareList, tracing.NewTracingMiddleware(logger))
	middlewareList = append(middlewareList, gerr.NewErrorMiddleware())
	if cfg.Services[serviceName].Metrics.ListenAddress != "" {
		middlewareList = metrics.NewMiddlewares(cfg.Services[serviceName].Metrics, middlewareList...)
	}
	return middlewareList, nil
}
