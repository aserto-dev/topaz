package middlewares

import (
	"context"

	"github.com/aserto-dev/aserto-grpc/grpcutil"
	"github.com/aserto-dev/aserto-grpc/grpcutil/middlewares/gerr"
	"github.com/aserto-dev/aserto-grpc/grpcutil/middlewares/request"
	"github.com/aserto-dev/aserto-grpc/grpcutil/middlewares/tracing"
	"github.com/aserto-dev/topaz/pkg/app/auth"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

func GetMiddlewaresForService(ctx context.Context, cfg *config.Config, logger *zerolog.Logger) ([]grpc.ServerOption, error) {
	var middlewareList grpcutil.Middlewares
	if len(cfg.Auth.APIKeys) > 0 {
		authmiddleware, err := auth.NewAPIKeyAuthMiddleware(ctx, &cfg.Auth, logger)
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
		NewTenantIDMiddleware(cfg),
		tracing.NewTracingMiddleware(logger),
		gerr.NewErrorMiddleware(),
	)

	var opts []grpc.ServerOption
	unary, stream := middlewareList.AsGRPCOptions()

	opts = append(opts, unary, stream)

	return opts, nil
}
