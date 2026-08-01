//nolint:testpackage
package impl

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestJWTResolverSecure(t *testing.T) {
	idToken := os.Getenv("TOPAZ_TEST_ID_TOKEN")
	if idToken == "" {
		t.Skipf("No id_token, env var TOPAZ_TEST_ID_TOKEN is not set")
	}

	resolver, err := NewJWTResolver(t.Context(), &config.JWT{
		AcceptableTimeSkewSeconds: 5,
		AllowedConfigurationEndpoints: []string{
			"https://trial-3441947-admin.okta.com/oauth2/default/.well-known/openid-configuration",
			"https://aserto.us.auth0.com/.well-known/openid-configuration",
		},
		CacheRefreshMinInterval: "5m",
		CacheRefreshMaxInterval: "15m",
	})
	require.NoError(t, err)

	if err := resolver.Start(t.Context()); err != nil {
		require.NoError(t, err)
	}

	subject, err := resolver.ResolveSubject(t.Context(), idToken)
	require.NoError(t, err)
	require.NotEmpty(t, subject)

	t.Logf("subject: %s", subject)

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	resolver.Stop(ctx)
}

func TestJWTResolverInsecure(t *testing.T) {
	idToken := os.Getenv("TOPAZ_TEST_ID_TOKEN")
	if idToken == "" {
		t.Skipf("No id_token, env var TOPAZ_TEST_ID_TOKEN is not set")
	}

	resolver, err := NewJWTResolver(t.Context(), &config.JWT{})
	require.NoError(t, err)

	if err := resolver.Start(t.Context()); err != nil {
		require.NoError(t, err)
	}

	subject, err := resolver.ResolveSubject(t.Context(), idToken)
	require.NoError(t, err)
	require.NotEmpty(t, subject)

	t.Logf("subject: %s", subject)

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	resolver.Stop(ctx)
}
