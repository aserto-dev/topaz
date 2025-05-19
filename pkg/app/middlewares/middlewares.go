package middlewares

import (
	"context"

	grpcutil "github.com/aserto-dev/aserto-grpc"
	"github.com/aserto-dev/aserto-grpc/middlewares/gerr"
	"github.com/aserto-dev/aserto-grpc/middlewares/request"
	"github.com/aserto-dev/aserto-grpc/middlewares/tracing"
	"github.com/aserto-dev/topaz/pkg/authentication"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"google.golang.org/grpc"
)

func GetMiddlewaresForService(ctx context.Context, cfg *config.Config, logger *zerolog.Logger) ([]grpc.ServerOption, error) {
	var middlewareList grpcutil.Middlewares

	if len(cfg.Auth.APIKeys) > 0 {
		v3Cfg := MigAuthnConfig(&cfg.Auth)
		middlewareList = append(middlewareList, authentication.NewMiddleware(&v3Cfg))
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

// TODO: remove when converting to native v3 config.
func MigAuthnConfig(v2 *config.AuthnConfig) authentication.Config {
	return authentication.Config{
		Enabled:  len(v2.Keys) != 0,
		Provider: authentication.LocalAuthenticationPlugin,
		Local: authentication.LocalConfig{
			Keys: v2.Keys,
			Options: authentication.CallOptions{
				Default: migAuthnOptions(&v2.Options.Default),
				Overrides: lo.Map(
					v2.Options.Overrides,
					func(override2 config.OptionOverrides, _ int) authentication.OptionOverrides {
						return authentication.OptionOverrides{
							Paths:    override2.Paths,
							Override: migAuthnOptions(&override2.Override),
						}
					},
				),
			},
		},
	}
}

func migAuthnOptions(v2 *config.Options) authentication.Options {
	return authentication.Options{
		AllowAnonymous: v2.EnableAnonymous || !v2.EnableAPIKey,
	}
}
