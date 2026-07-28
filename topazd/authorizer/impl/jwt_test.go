//nolint:testpackage
package impl

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jwx-go/jwkfetch/v4"
	"github.com/lestrrat-go/httprc/v3"
	"github.com/lestrrat-go/jwx/v4/jwa"
	"github.com/lestrrat-go/jwx/v4/jwk"
	"github.com/lestrrat-go/jwx/v4/jws"
	"github.com/lestrrat-go/jwx/v4/jwt"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/aserto-dev/topaz/pkg/config"
)

const (
	testKid     = "test-kid"
	testSubject = "test-subject"
)

// jwksTestServer serves a JWK set at /.well-known/jwks.json and counts
// how many times it has been fetched, so tests can assert on caching
// behavior. It has no /.well-known/openid-configuration handler, so
// jwksURL() falls back to the jwks.json well-known path, matching how a
// bare JWKS-hosting issuer behaves.
type jwksTestServer struct {
	*httptest.Server

	hits atomic.Int64
}

// newECDSAKey generates a fresh ECDSA P-256 key pair for signing test tokens.
func newECDSAKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	return priv
}

// newJWKSTestServer publishes priv's public key (with kid/alg set, as
// required by jwt.WithKeySet) as the JWK set served at the issuer's JWKS
// endpoint.
func newJWKSTestServer(t *testing.T, priv *ecdsa.PrivateKey) *jwksTestServer {
	t.Helper()

	pub, err := jwk.PublicKeyOf(priv)
	require.NoError(t, err)
	require.NoError(t, pub.Set(jwk.KeyIDKey, testKid))
	require.NoError(t, pub.Set(jwk.AlgorithmKey, jwa.ES256()))

	set := jwk.NewSet()
	require.NoError(t, set.AddKey(pub))

	buf, err := json.Marshal(set)
	require.NoError(t, err)

	s := &jwksTestServer{}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/jwks.json", func(w http.ResponseWriter, r *http.Request) {
		s.hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(buf)
	})

	s.Server = httptest.NewServer(mux)
	t.Cleanup(s.Close)

	return s
}

// signTestToken builds and signs a JWT for testSubject with the given
// issuer and expiry, using kid testKid so it matches the JWK published in
// the test server's key set.
func signTestToken(t *testing.T, priv *ecdsa.PrivateKey, issuer string, iat, exp time.Time) string {
	t.Helper()

	tok, err := jwt.NewBuilder().
		Issuer(issuer).
		Subject(testSubject).
		IssuedAt(iat).
		Expiration(exp).
		Build()
	require.NoError(t, err)

	hdrs := jws.NewHeaders()
	require.NoError(t, hdrs.Set(jws.KeyIDKey, testKid))

	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.ES256(), priv, jws.WithProtectedHeaders(hdrs)))
	require.NoError(t, err)

	return string(signed)
}

func newTestAuthorizerServer(t *testing.T) *AuthorizerServer {
	t.Helper()

	logger := zerolog.Nop()

	jwkCache, err := jwkfetch.NewCache(context.Background(), httprc.NewClient())
	require.NoError(t, err)

	t.Cleanup(func() { _ = jwkCache.Shutdown(context.Background()) })

	return &AuthorizerServer{
		cfg:      &config.Common{},
		logger:   &logger,
		jwkCache: jwkCache,
	}
}

func TestGetIdentityFromJWT_Valid(t *testing.T) {
	t.Parallel()

	priv := newECDSAKey(t)
	server := newJWKSTestServer(t, priv)

	s := newTestAuthorizerServer(t)

	now := time.Now()
	token := signTestToken(t, priv, server.URL, now, now.Add(time.Hour))

	ident, err := s.getIdentityFromJWT(context.Background(), token)
	require.NoError(t, err)
	require.Equal(t, testSubject, ident)
}

// TestIssuerAllowed_ExactAndAnchoredPrefix checks that issuerAllowed only
// treats a configured entry as a path prefix when it ends in "/", and that
// the prefix match is anchored at that "/" boundary so it can't be defeated
// by an attacker-chosen suffix such as "https://idp.example.com.attacker.evil".
func TestIssuerAllowed_ExactAndAnchoredPrefix(t *testing.T) {
	t.Parallel()

	const (
		baseIssuer   = "https://idp.example.com"
		attackIssuer = baseIssuer + ".attacker.evil"
	)

	cases := []struct {
		name    string
		allowed []string
		issuer  string
		want    bool
	}{
		{"exact match", []string{baseIssuer}, baseIssuer, true},
		{"unrelated issuer", []string{baseIssuer}, "https://other.example.com", false},
		{"suffix attack against unslashed entry", []string{baseIssuer}, attackIssuer, false},
		{"anchored prefix matches sub-path", []string{baseIssuer + "/"}, baseIssuer + "/tenant-a", true},
		{"anchored prefix rejects suffix attack", []string{baseIssuer + "/"}, attackIssuer, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, issuerAllowed(tc.allowed, tc.issuer))
		})
	}
}

func TestGetIdentityFromJWT_AcceptedIssuer_Allowed(t *testing.T) {
	t.Parallel()

	priv := newECDSAKey(t)
	server := newJWKSTestServer(t, priv)

	s := newTestAuthorizerServer(t)
	s.cfg.JWT.AllowedIssuers = []string{server.URL, "https://other-issuer.example"}

	now := time.Now()
	token := signTestToken(t, priv, server.URL, now, now.Add(time.Hour))

	ident, err := s.getIdentityFromJWT(context.Background(), token)
	require.NoError(t, err)
	require.Equal(t, testSubject, ident)
}

func TestGetIdentityFromJWT_AcceptedIssuer_Rejected(t *testing.T) {
	t.Parallel()

	priv := newECDSAKey(t)
	server := newJWKSTestServer(t, priv)

	s := newTestAuthorizerServer(t)
	s.cfg.JWT.AllowedIssuers = []string{"https://other-issuer.example"}

	now := time.Now()
	token := signTestToken(t, priv, server.URL, now, now.Add(time.Hour))

	_, err := s.getIdentityFromJWT(context.Background(), token)
	require.Error(t, err)

	// The issuer check must happen before any JWKS discovery/fetch, so a
	// rejected issuer never causes an outbound request to be made.
	require.Equal(t, int64(0), server.hits.Load(),
		"expected no JWKS fetch for a non-accepted issuer")
}

func TestGetIdentityFromJWT_CachesJWKSPerIssuer(t *testing.T) {
	t.Parallel()

	priv := newECDSAKey(t)
	server := newJWKSTestServer(t, priv)

	s := newTestAuthorizerServer(t)

	now := time.Now()

	for i := range 3 {
		token := signTestToken(t, priv, server.URL, now, now.Add(time.Hour))

		_, err := s.getIdentityFromJWT(context.Background(), token)
		require.NoErrorf(t, err, "call %d", i)
	}

	// registerJWKSURL only registers+refreshes on the first call for a
	// given issuer; subsequent calls should hit the already-registered
	// cache entry via Lookup and not re-fetch the JWKS endpoint.
	require.LessOrEqual(t, server.hits.Load(), int64(2),
		"expected the JWKS endpoint to be fetched only on first registration, not on every call")
}

func TestGetIdentityFromJWT_Expired(t *testing.T) {
	t.Parallel()

	priv := newECDSAKey(t)
	server := newJWKSTestServer(t, priv)

	s := newTestAuthorizerServer(t)

	past := time.Now().Add(-2 * time.Hour)
	token := signTestToken(t, priv, server.URL, past, past.Add(time.Hour))

	_, err := s.getIdentityFromJWT(context.Background(), token)
	require.Error(t, err)
}

func TestGetIdentityFromJWT_WrongSigningKey(t *testing.T) {
	t.Parallel()

	server := newJWKSTestServer(t, newECDSAKey(t))

	// Sign with a different key than the one published in the JWKS.
	otherPriv := newECDSAKey(t)

	s := newTestAuthorizerServer(t)

	now := time.Now()
	token := signTestToken(t, otherPriv, server.URL, now, now.Add(time.Hour))

	_, err := s.getIdentityFromJWT(context.Background(), token)
	require.Error(t, err)
}

func TestGetIdentityFromJWT_Malformed(t *testing.T) {
	t.Parallel()

	s := newTestAuthorizerServer(t)

	_, err := s.getIdentityFromJWT(context.Background(), "not-a-jwt")
	require.Error(t, err)
}

func TestGetIdentityFromJWT_NoSchemeIssuer(t *testing.T) {
	t.Parallel()

	priv := newECDSAKey(t)

	s := newTestAuthorizerServer(t)

	now := time.Now()
	token := signTestToken(t, priv, "no-scheme-issuer", now, now.Add(time.Hour))

	_, err := s.getIdentityFromJWT(context.Background(), token)
	require.Error(t, err)
}

func TestJwksURLFromCache_CachesPerIssuer(t *testing.T) {
	t.Parallel()

	server := newJWKSTestServer(t, newECDSAKey(t))

	s := newTestAuthorizerServer(t)

	url1, err := s.jwksURLFromCache(context.Background(), server.URL)
	require.NoError(t, err)

	url2, err := s.jwksURLFromCache(context.Background(), server.URL)
	require.NoError(t, err)

	require.Equal(t, url1, url2)
	// jwksURLFromCache does the well-known discovery HTTP call itself
	// (separate from the JWKS fetch counted by jwksTestServer); the
	// discovery result is cached in s.issuers, so a second call for the
	// same issuer must not repeat the discovery request.
	require.Equal(t, int64(0), server.hits.Load(),
		"discovery should not touch the jwks.json endpoint")
}
