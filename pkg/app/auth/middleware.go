package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aserto-dev/aserto-grpc/authn"
	"github.com/aserto-dev/aserto-grpc/authn/apikey"
	authn_config "github.com/aserto-dev/aserto-grpc/authn/config"
	"github.com/aserto-dev/aserto-grpc/grpcutil"
	"github.com/aserto-dev/go-utils/cerr"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

type APIKeyAuthMiddleware struct {
	apiAuth *apikey.Authenticator

	cfg    *authn_config.Config
	logger *zerolog.Logger
}

func NewAPIKeyAuthMiddleware(
	ctx context.Context,
	cfg *authn_config.Config,
	logger *zerolog.Logger,
) (*APIKeyAuthMiddleware, error) {

	apiAuth := apikey.New(cfg.APIKeys)

	return &APIKeyAuthMiddleware{
		apiAuth: apiAuth,
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
			httpAccountIDHeader(r),
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
	path, authHeader, accountIDOverride string,
) (context.Context, error) {
	tenantID := grpcutil.ExtractTenantID(ctx)

	options := a.cfg.Options.ForPath(path)
	if options.NeedsTenant && tenantID == "" {
		return ctx, cerr.ErrNoTenantID
	}

	basicAPIKey, err := authn.ParseAuthHeader(authHeader, "basic")
	if err != nil {
		a.logger.Trace().Err(err).Str("auth_header", authHeader).Msg("failed to parse basic auth header")
	}

	if basicAPIKey != "" {
		if options.EnableAPIKey {
			if accountID, err := a.apiAuth.Authenticate(basicAPIKey, accountIDOverride); err == nil {
				return context.WithValue(ctx, grpcutil.HeaderAsertoAccountID, accountID), nil
			}
			a.logger.Debug().AnErr("error", err).Str("api_key", basicAPIKey).Msg("authentication failed")
		}
	}

	if options.EnableAnonymous {
		return ctx, nil
	}

	return ctx, cerr.ErrAuthenticationFailed
}

func (a *APIKeyAuthMiddleware) grpcAuthenticate(ctx context.Context) (context.Context, error) {
	method, _ := grpc.Method(ctx)
	return a.authenticate(ctx, method, grpcAuthHeader(ctx), grpcAccountIDHeader(ctx))
}

func grpcAuthHeader(ctx context.Context) string {
	return metautils.ExtractIncoming(ctx).Get("Authorization")
}

func grpcAccountIDHeader(ctx context.Context) string {
	return metautils.ExtractIncoming(ctx).Get(string(grpcutil.HeaderAsertoAccountID))
}

func httpAuthHeader(r *http.Request) string {
	return r.Header.Get("Authorization")
}

func httpAccountIDHeader(r *http.Request) string {
	return r.Header.Get(string(grpcutil.HeaderAsertoAccountID))
}
