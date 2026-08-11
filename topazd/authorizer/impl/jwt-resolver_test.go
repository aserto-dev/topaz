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

	"github.com/lestrrat-go/jwx/v4/jwa"
	"github.com/lestrrat-go/jwx/v4/jwk"
	"github.com/lestrrat-go/jwx/v4/jws"
	"github.com/lestrrat-go/jwx/v4/jwt"
	"github.com/stretchr/testify/require"

	"github.com/aserto-dev/topaz/pkg/config"
)

// ---------------------------------------------------------------------------
// Local (in-process) coverage of the secure resolution path.
//
// The tests below replace the coverage that used to live
// in the deleted jwt_test.go, adapted to the OIDC-discovery-based resolver:
// jwtResolver.Start() now requires a reachable ".well-known/openid-configuration"
// document per allowed issuer (fetched via OidcClient), not just a JWKS
// endpoint, so the mock test server has to serve both.
//
// OidcClient's default transport (NewOidcClient) deliberately refuses
// loopback/private addresses and requires https as an SSRF guard, so it can
// never reach an in-process httptest server. jwtResolver.oidcClient is an
// unexported field precisely so tests in this package can substitute an
// OidcClient built around httptest's own TLS-trusting client for the
// discovery hop; the JWKS fetch itself goes through jwkfetch/httprc (used
// via r.cache), which has no such restriction and is exercised the same way
// prior tests exercised it, over plain HTTP.
// ---------------------------------------------------------------------------

const (
	testKid     = "test-kid"
	testSubject = "test-subject"
)

// mockOIDCProvider is a local, in-process issuer serving both endpoints the
// new resolver needs: OIDC discovery (TLS, matching OidcClient's https
// requirement) and the JWKS document it points to (plain HTTP, matching how
// jwkfetch/httprc is exercised elsewhere in this package).
type mockOIDCProvider struct {
	discovery *httptest.Server
	jwks      *httptest.Server

	discoveryHits atomic.Int64
	jwksHits      atomic.Int64
}

func newMockOIDCProvider(t *testing.T, priv *ecdsa.PrivateKey) *mockOIDCProvider {
	t.Helper()

	pub, err := jwk.PublicKeyOf(priv)
	require.NoError(t, err)
	require.NoError(t, pub.Set(jwk.KeyIDKey, testKid))
	require.NoError(t, pub.Set(jwk.AlgorithmKey, jwa.ES256()))

	set := jwk.NewSet()
	require.NoError(t, set.AddKey(pub))

	jwksBuf, err := json.Marshal(set)
	require.NoError(t, err)

	p := &mockOIDCProvider{}

	jwksMux := http.NewServeMux()
	jwksMux.HandleFunc("/jwks.json", func(w http.ResponseWriter, _ *http.Request) {
		p.jwksHits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jwksBuf)
	})

	p.jwks = httptest.NewServer(jwksMux)
	t.Cleanup(p.jwks.Close)

	discoveryMux := http.NewServeMux()
	discoveryMux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		p.discoveryHits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"issuer":   p.Issuer(),
			"jwks_uri": p.jwks.URL + "/jwks.json",
		})
	})

	p.discovery = httptest.NewTLSServer(discoveryMux)
	t.Cleanup(p.discovery.Close)

	return p
}

// Issuer returns the issuer URL to use both as the configured
// config.jwt.allowed_issuers entry and as the "iss" claim of test tokens.
func (p *mockOIDCProvider) Issuer() string {
	return p.discovery.URL
}

// newECDSAKey generates a fresh ECDSA P-256 key pair for signing test tokens.
func newECDSAKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	return priv
}

// signTestToken builds and signs a JWT for testSubject with the given issuer
// and expiry, using kid testKid so it matches the JWK published by
// newMockOIDCProvider. audience is omitted when empty.
func signTestToken(t *testing.T, priv *ecdsa.PrivateKey, issuer, audience string, iat, exp time.Time) string {
	t.Helper()

	builder := jwt.NewBuilder().
		Issuer(issuer).
		Subject(testSubject).
		IssuedAt(iat).
		Expiration(exp)

	if audience != "" {
		builder = builder.Audience([]string{audience})
	}

	tok, err := builder.Build()
	require.NoError(t, err)

	hdrs := jws.NewHeaders()
	require.NoError(t, hdrs.Set(jws.KeyIDKey, testKid))

	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.ES256(), priv, jws.WithProtectedHeaders(hdrs)))
	require.NoError(t, err)

	return string(signed)
}

// newLocalTestResolver builds a jwtResolver wired to provider's discovery
// endpoint and starts it, registering cleanup to stop it.
func newLocalTestResolver(t *testing.T, jwtConfig *config.JWT, provider *mockOIDCProvider) *jwtResolver {
	t.Helper()

	resolver, err := NewJWTResolver(t.Context(), jwtConfig)
	require.NoError(t, err)

	// Substitute an OidcClient built around httptest's TLS-trusting client so
	// discovery can reach the in-process, loopback-bound mock server. See the
	// package comment above for why this is necessary.
	resolver.oidcClient = &OidcClient{httpClient: provider.discovery.Client()}

	require.NoError(t, resolver.Start(t.Context()))

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = resolver.Stop(ctx)
	})

	return resolver
}

func TestJWTResolverLocal_Valid(t *testing.T) {
	t.Parallel()

	priv := newECDSAKey(t)
	provider := newMockOIDCProvider(t, priv)

	resolver := newLocalTestResolver(t, &config.JWT{
		AcceptableTimeSkewSeconds: 5,
		AllowedIssuers:            []string{provider.Issuer()},
	}, provider)

	now := time.Now()
	token := signTestToken(t, priv, provider.Issuer(), "", now, now.Add(time.Hour))

	subject, err := resolver.ResolveSubject(t.Context(), token)
	require.NoError(t, err)
	require.Equal(t, testSubject, subject)
}

func TestJWTResolverLocal_UnknownIssuerRejected(t *testing.T) {
	t.Parallel()

	allowedPriv := newECDSAKey(t)
	allowedProvider := newMockOIDCProvider(t, allowedPriv)

	otherPriv := newECDSAKey(t)
	otherProvider := newMockOIDCProvider(t, otherPriv)

	resolver := newLocalTestResolver(t, &config.JWT{
		AcceptableTimeSkewSeconds: 5,
		AllowedIssuers:            []string{allowedProvider.Issuer()},
	}, allowedProvider)

	now := time.Now()
	token := signTestToken(t, otherPriv, otherProvider.Issuer(), "", now, now.Add(time.Hour))

	_, err := resolver.ResolveSubject(t.Context(), token)
	require.Error(t, err)

	// A non-allowed issuer must never be contacted: Start() only registers
	// issuers listed in config.jwt.allowed_issuers, and ResolveSubject must
	// reject an unrecognized "iss" claim via the issuerToConfig map lookup
	// before attempting any discovery or JWKS fetch against it.
	require.Equal(t, int64(0), otherProvider.discoveryHits.Load(),
		"expected no discovery request for a non-allowed issuer")
	require.Equal(t, int64(0), otherProvider.jwksHits.Load(),
		"expected no JWKS fetch for a non-allowed issuer")
}

func TestJWTResolverLocal_DiscoveryAndJWKSAreCached(t *testing.T) {
	t.Parallel()

	priv := newECDSAKey(t)
	provider := newMockOIDCProvider(t, priv)

	resolver := newLocalTestResolver(t, &config.JWT{
		AcceptableTimeSkewSeconds: 5,
		AllowedIssuers:            []string{provider.Issuer()},
	}, provider)

	require.Equal(t, int64(1), provider.discoveryHits.Load(),
		"discovery should happen exactly once, during Start")

	now := time.Now()

	for i := range 3 {
		token := signTestToken(t, priv, provider.Issuer(), "", now, now.Add(time.Hour))

		_, err := resolver.ResolveSubject(t.Context(), token)
		require.NoErrorf(t, err, "call %d", i)
	}

	require.Equal(t, int64(1), provider.discoveryHits.Load(),
		"discovery must not be repeated on every ResolveSubject call")
	require.LessOrEqual(t, provider.jwksHits.Load(), int64(2),
		"expected the JWKS endpoint to be served from the background-refreshed cache, not re-fetched on every call")
}

func TestJWTResolverLocal_Expired(t *testing.T) {
	t.Parallel()

	priv := newECDSAKey(t)
	provider := newMockOIDCProvider(t, priv)

	resolver := newLocalTestResolver(t, &config.JWT{
		AcceptableTimeSkewSeconds: 5,
		AllowedIssuers:            []string{provider.Issuer()},
	}, provider)

	past := time.Now().Add(-2 * time.Hour)
	token := signTestToken(t, priv, provider.Issuer(), "", past, past.Add(time.Hour))

	_, err := resolver.ResolveSubject(t.Context(), token)
	require.Error(t, err)
}

func TestJWTResolverLocal_WrongSigningKey(t *testing.T) {
	t.Parallel()

	priv := newECDSAKey(t)
	provider := newMockOIDCProvider(t, priv)

	resolver := newLocalTestResolver(t, &config.JWT{
		AcceptableTimeSkewSeconds: 5,
		AllowedIssuers:            []string{provider.Issuer()},
	}, provider)

	// Sign with a different key than the one published in the JWKS.
	otherPriv := newECDSAKey(t)

	now := time.Now()
	token := signTestToken(t, otherPriv, provider.Issuer(), "", now, now.Add(time.Hour))

	_, err := resolver.ResolveSubject(t.Context(), token)
	require.Error(t, err)
}

func TestJWTResolverLocal_Malformed(t *testing.T) {
	t.Parallel()

	priv := newECDSAKey(t)
	provider := newMockOIDCProvider(t, priv)

	resolver := newLocalTestResolver(t, &config.JWT{
		AcceptableTimeSkewSeconds: 5,
		AllowedIssuers:            []string{provider.Issuer()},
	}, provider)

	_, err := resolver.ResolveSubject(t.Context(), "not-a-jwt")
	require.Error(t, err)
}

func TestJWTResolverLocal_ExpectedAudience_Match(t *testing.T) {
	t.Parallel()

	priv := newECDSAKey(t)
	provider := newMockOIDCProvider(t, priv)

	resolver := newLocalTestResolver(t, &config.JWT{
		AcceptableTimeSkewSeconds: 5,
		AllowedIssuers:            []string{provider.Issuer()},
		ExpectedAudience:          "expected-audience",
	}, provider)

	now := time.Now()
	token := signTestToken(t, priv, provider.Issuer(), "expected-audience", now, now.Add(time.Hour))

	subject, err := resolver.ResolveSubject(t.Context(), token)
	require.NoError(t, err)
	require.Equal(t, testSubject, subject)
}

func TestJWTResolverLocal_ExpectedAudience_Mismatch(t *testing.T) {
	t.Parallel()

	priv := newECDSAKey(t)
	provider := newMockOIDCProvider(t, priv)

	resolver := newLocalTestResolver(t, &config.JWT{
		AcceptableTimeSkewSeconds: 5,
		AllowedIssuers:            []string{provider.Issuer()},
		ExpectedAudience:          "expected-audience",
	}, provider)

	now := time.Now()
	token := signTestToken(t, priv, provider.Issuer(), "wrong-audience", now, now.Add(time.Hour))

	_, err := resolver.ResolveSubject(t.Context(), token)
	require.Error(t, err)
}

// TestJWTResolverLocal_InsecureMode documents the intended behavior of
// resolveSubjectInsecure: when config.jwt.allowed_issuers is empty, there is
// no JWKS trust anchor to verify a signature against, so ResolveSubject
// deliberately skips signature verification and only applies the
// validations jwt.WithValidate(true) performs without a keyset (exp/nbf/iat).
func TestJWTResolverLocal_InsecureMode(t *testing.T) {
	t.Parallel()

	resolver, err := NewJWTResolver(t.Context(), &config.JWT{
		AcceptableTimeSkewSeconds: 5,
		// AllowedIssuers intentionally left empty/unset - the shipped default.
	})
	require.NoError(t, err)
	require.NoError(t, resolver.Start(t.Context()))

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = resolver.Stop(ctx)
	})

	key := newECDSAKey(t)

	t.Run("well-formed token resolves without signature verification", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		token := signTestToken(t, key, "https://issuer.example/not-configured", "", now, now.Add(time.Hour))

		subject, err := resolver.ResolveSubject(t.Context(), token)
		require.NoError(t, err)
		require.Equal(t, testSubject, subject)
	})

	t.Run("expired token is still rejected", func(t *testing.T) {
		t.Parallel()

		past := time.Now().Add(-2 * time.Hour)
		token := signTestToken(t, key, "https://issuer.example/not-configured", "", past, past.Add(time.Hour))

		_, err := resolver.ResolveSubject(t.Context(), token)
		require.Error(t, err)
	})
}
