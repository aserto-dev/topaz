package impl_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/topazd/authorizer/impl"
)

// NOTE: This file contains Development tests for interactive debugging of resolver flow.
//       These tests are excluded from the tests suites

func TestJWTResolverSecure(t *testing.T) {
	if os.Getenv("TOPAZ_DEBUG") != "1" {
		t.Skip()
	}

	idToken := os.Getenv("TOPAZ_TEST_ID_TOKEN")
	if idToken == "" {
		t.Skipf("No id_token, env var TOPAZ_TEST_ID_TOKEN is not set")
	}

	resolver, err := impl.NewJWTResolver(t.Context(), &config.JWT{
		AcceptableTimeSkewSeconds: 5,
		AllowedIssuers: []string{
			"https://trial-3441947.okta.com/oauth2/default",
			"https://aserto.us.auth0.com/",
		},
		CacheRefreshMinInterval: "5m",
		CacheRefreshMaxInterval: "15m",
		ExpectedAudience:        "98ofxNoUdgVu7vuYAddWW2WpglFM4til",
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

	_ = resolver.Stop(ctx)
}

func TestJWTResolverInsecure(t *testing.T) {
	if os.Getenv("TOPAZ_DEBUG") != "1" {
		t.Skip()
	}

	idToken := os.Getenv("TOPAZ_TEST_ID_TOKEN")
	if idToken == "" {
		t.Skipf("No id_token, env var TOPAZ_TEST_ID_TOKEN is not set")
	}

	resolver, err := impl.NewJWTResolver(t.Context(), &config.JWT{})
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

	_ = resolver.Stop(ctx)
}
