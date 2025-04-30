package authentication

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	"github.com/aserto-dev/topaz/pkg/app/handlers"
	"github.com/aserto-dev/topaz/pkg/middleware"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"google.golang.org/grpc"
)

type Middleware struct {
	// keys maps API keys to their associated identities.
	keys    keyset
	options CallOptions
}

//nolint:ireturn  // Factory function.
func New(cfg *Config) middleware.Server {
	if !cfg.Enabled {
		return middleware.Noop{}
	}

	return NewMiddleware(cfg)
}

func NewMiddleware(cfg *Config) *Middleware {
	keys := lo.SliceToMap(cfg.Local.Keys, func(key string) (string, struct{}) {
		return key, struct{}{}
	})

	return &Middleware{keys, cfg.Local.Options}
}

type keyset map[string]struct{}

func (a *Middleware) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		newCtx, err := a.grpcAuthenticate(ctx)
		if err != nil {
			return nil, err
		}

		return handler(newCtx, req)
	}
}

func (a *Middleware) Stream() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
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

func (a *Middleware) Handler(next http.Handler) http.Handler {
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

func (a *Middleware) authenticate(
	ctx context.Context,
	path, authHeader string,
) (context.Context, error) {
	if len(a.keys) == 0 {
		return ctx, nil
	}

	options := a.options.ForPath(path)
	if options.AllowAnonymous {
		return ctx, nil
	}

	basicAPIKey, err := parseAuthHeader(authHeader, "basic")
	if err != nil {
		zerolog.Ctx(ctx).Trace().Err(err).Str("auth_header", authHeader).Msg("failed to parse basic auth header")
	}

	// allow the request if the API key is present in the config
	if _, ok := a.keys[basicAPIKey]; ok {
		return ctx, nil
	}

	return ctx, aerr.ErrAuthenticationFailed
}

func (a *Middleware) grpcAuthenticate(ctx context.Context) (context.Context, error) {
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
	scheme, header, ok := strings.Cut(val, " ")
	if !ok {
		return "", aerr.ErrAuthenticationFailed.Msg("Bad authorization string")
	}

	if !strings.EqualFold(scheme, expectedScheme) {
		return "", aerr.ErrAuthenticationFailed.Msgf("Request unauthenticated with expected scheme, expected: %s", expectedScheme)
	}

	return header, nil
}

// TODO: Fold this code path into the normal 'authenticate' function.
func (a *Middleware) ConfigHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		options := a.options.ForPath(r.URL.Path)
		if options.AllowAnonymous {
			ctx := context.WithValue(r.Context(), handlers.AuthenticatedUser, true)
			h.ServeHTTP(w, r.WithContext(ctx))

			return
		}

		// if we reached this point, an API key is required
		ctx := context.WithValue(r.Context(), handlers.AuthEnabled, true)

		authHeader := httpAuthHeader(r)
		if authHeader == "" {
			// auth header is not present =>  the user is unauthenticated and did not provide a token
			ctx = context.WithValue(ctx, handlers.AuthenticatedUser, false)
			h.ServeHTTP(w, r.WithContext(ctx))

			return
		}

		basicAPIKey, err := parseAuthHeader(authHeader, "basic")
		if err != nil {
			returnStatusUnauthorized(w, "Invalid authorization header. expected 'basic' scheme.", zerolog.Ctx(r.Context()))

			return
		}

		if _, ok := a.keys[basicAPIKey]; ok {
			ctx = context.WithValue(ctx, handlers.AuthenticatedUser, true)
			h.ServeHTTP(w, r.WithContext(ctx))

			return
		}

		// the user is not authenticated because the key they provided is incorrect
		returnStatusUnauthorized(w, "The API key is invalid.", zerolog.Ctx(r.Context()))
	})
}

func returnStatusUnauthorized(w http.ResponseWriter, errMsg string, log *zerolog.Logger) {
	w.WriteHeader(http.StatusUnauthorized)

	if _, err := w.Write([]byte(errMsg)); err != nil {
		log.Error().Err(err).Msg("could not write response message")
	}
}
