//nolint:testpackage
package impl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	oktaIssuer  = "https://trial-3441947.okta.com/oauth2/default"
	auth0Issuer = "https://aserto.us.auth0.com/"
)

func TestOidcClient(t *testing.T) {
	ctx := t.Context()

	issuers := []string{auth0Issuer, oktaIssuer}

	client := NewOidcClient()

	for _, issuer := range issuers {
		config, err := client.FetchAndValidateConfig(ctx, issuer)
		require.NoError(t, err)

		t.Logf("Validated Issuer: %s\n", config.Issuer)
		t.Logf("Validated JWKS URI: %s\n", config.JwksURI)
	}
}
