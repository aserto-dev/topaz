package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

type APIKeyAuthMiddleware struct {
	apiAuth map[string]string

	// TODO: figure out what we want to do with the config.
	cfg    *config.AuthnConfig
	logger *zerolog.Logger
}

func NewAPIKeyAuthMiddleware(
	ctx context.Context,
	// TODO: figure out what we want to do with the config.
	cfg *config.AuthnConfig,
	logger *zerolog.Logger,
) (*APIKeyAuthMiddleware, error) {
	return &APIKeyAuthMiddleware{
		apiAuth: cfg.APIKeys,
		cfg:     cfg,
		logger:  logger,
	}, nil
}

func (a *APIKeyAuthMiddleware) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx, err := a.grpcAuthenticate(ctx)
		if err != nil {
			return nil, err
		}

		return handler(newCtx, req)
	}
}

func (a *APIKeyAuthMiddleware) Stream() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()

		newCtx, err := a.grpcAuthenticate(ctx)
		if err != nil {
			return err
		}

		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = newCtx

		return handler(srv, wrapped)
	}
}

func (a *APIKeyAuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newCtx, err := a.authenticate(
			r.Context(),
			r.URL.Path,
			httpAuthHeader(r),
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("%q", err.Error()), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r.WithContext(newCtx))
	})
}

func (a *APIKeyAuthMiddleware) authenticate(
	ctx context.Context,
	path, authHeader string,
) (context.Context, error) {
	options := a.cfg.Options.ForPath(path)

	if options.EnableAnonymous {
		return ctx, nil
	}

	// if no API keys are defined or EnableAPIKey is not set, allow the request
	if (len(a.cfg.APIKeys) == 0) || !options.EnableAPIKey {
		return ctx, nil
	}

	basicAPIKey, err := parseAuthHeader(authHeader, "basic")
	if err != nil {
		a.logger.Trace().Err(err).Str("auth_header", authHeader).Msg("failed to parse basic auth header")
	}

	// allow the request if the API key is present in the config
	if _, ok := a.cfg.APIKeys[basicAPIKey]; ok {
		return ctx, nil
	}

	return ctx, aerr.ErrAuthenticationFailed
}

func (a *APIKeyAuthMiddleware) grpcAuthenticate(ctx context.Context) (context.Context, error) {
	method, _ := grpc.Method(ctx)
	return a.authenticate(ctx, method, grpcAuthHeader(ctx))
}

func grpcAuthHeader(ctx context.Context) string {
	return metautils.ExtractIncoming(ctx).Get("Authorization")
}

func httpAuthHeader(r *http.Request) string {
	return r.Header.Get("Authorization")
}

func parseAuthHeader(val, expectedScheme string) (string, error) {
	splits := strings.SplitN(val, " ", 2)
	if len(splits) < 2 {
		return "", aerr.ErrAuthenticationFailed.Msg("Bad authorization string")
	}
	if !strings.EqualFold(splits[0], expectedScheme) {
		return "", aerr.ErrAuthenticationFailed.Msgf("Request unauthenticated with expected scheme, expected: %s", expectedScheme)
	}
	return splits[1], nil
}
