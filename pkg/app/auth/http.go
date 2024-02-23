package auth

import (
	"context"
	"net/http"

	"github.com/aserto-dev/topaz/pkg/app/handlers"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/rs/zerolog"
)

func (a *APIKeyAuthMiddleware) ConfigAuth(h http.Handler, authCfg config.AuthnConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if no API keys are defined or EnableAPIKey is not set, allow the request
		options := authCfg.Options.ForPath(r.URL.Path)

		if options.EnableAnonymous {
			ctx := context.WithValue(r.Context(), handlers.AuthenticatedUser, true)
			h.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		if (len(authCfg.APIKeys) == 0) || !options.EnableAPIKey {
			ctx := context.WithValue(r.Context(), handlers.AuthenticatedUser, true)
			h.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// if we reached this point, auth is enabled
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
			returnStatusUnauthorized(w, "Invalid authorization header. expected 'basic' scheme.", a.logger)
			return
		}

		if _, ok := authCfg.APIKeys[basicAPIKey]; ok {
			ctx = context.WithValue(ctx, handlers.AuthenticatedUser, true)
			h.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// the user is not authenticated because the key they provided is incorrect
		returnStatusUnauthorized(w, "The API key is invalid.", a.logger)
	})
}

func returnStatusUnauthorized(w http.ResponseWriter, errMsg string, log *zerolog.Logger) {
	w.WriteHeader(http.StatusUnauthorized)
	_, err := w.Write([]byte(errMsg))
	if err != nil {
		log.Error().Err(err).Msg("could not write response message")
	}
}
