package middlewares

import (
	"context"

	grpcutil "github.com/aserto-dev/topaz/internal/grpc"
	"github.com/aserto-dev/topaz/internal/grpc/middlewares/gerr"
	"github.com/aserto-dev/topaz/internal/grpc/middlewares/request"
	"github.com/aserto-dev/topaz/internal/grpc/middlewares/tracing"
	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/topazd/authentication"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

func GetMiddlewaresForService(ctx context.Context, cfg *config.Config, logger *zerolog.Logger) ([]grpc.ServerOption, error) {
	var middlewareList grpcutil.Middlewares

	if len(cfg.Auth.APIKeys) > 0 {
		authmiddleware, err := authentication.NewAPIKeyAuthMiddleware(ctx, &cfg.Auth, logger)
		if err != nil {
			return nil, err
		}

		middlewareList = append(middlewareList, authmiddleware)
	}

	// only attach policy instance information if discovery resource is configured.
	if cfg.OPA.Config.Discovery != nil && cfg.OPA.Config.Discovery.Resource != nil {
		middlewareList = append(middlewareList, NewInstanceMiddleware(cfg, logger))
	}

	// get tenant id from opa instance id.
	middlewareList = append(middlewareList,
		request.NewRequestIDMiddleware(),
		tracing.NewTracingMiddleware(logger),
		gerr.NewErrorMiddleware(),
	)

	unary, stream := middlewareList.AsGRPCOptions()
	opts := []grpc.ServerOption{unary, stream}

	return opts, nil
}
