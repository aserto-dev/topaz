package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/jwx-go/jwkfetch/v4"
	"github.com/lestrrat-go/httprc/v3"
	"github.com/lestrrat-go/jwx/v4/jwt"
)

type jwtResolver struct {
	cache          *jwkfetch.Cache
	jwtConfig      *config.JWT
	issuerToConfig map[string]*OIDCConfiguration
}

func NewJWTResolver(ctx context.Context, config *config.JWT) (*jwtResolver, error) {
	// create cache instance
	cache, err := jwkfetch.NewCache(ctx, httprc.NewClient())
	if err != nil {
		return nil, err
	}

	resolver := jwtResolver{
		cache:          cache,
		jwtConfig:      config,
		issuerToConfig: make(map[string]*OIDCConfiguration),
	}

	return &resolver, nil
}

func (r *jwtResolver) Start(ctx context.Context) error {
	// hydrate the configs map for each OIDC config URL registered in config.jwt.allowed_configuration_endpoints.
	for _, configURL := range r.jwtConfig.AllowedConfigurationEndpoints {
		config, err := fetchOIDCConfig(ctx, configURL)
		if err != nil {
			return err
		}

		r.issuerToConfig[config.Issuer] = config
	}

	// register each endpoints JWKS URI.
	for _, cfg := range r.issuerToConfig {
		err := r.cache.Register(
			ctx,
			cfg.JWKSURI,
			jwkfetch.WithWaitReady(true),
			jwkfetch.WithMinInterval(r.jwtConfig.CacheRefreshMinIntervalDuration()),
			jwkfetch.WithMaxInterval(r.jwtConfig.CacheRefreshMaxIntervalDuration()),
		)
		if err != nil {
			return err
		}
	}

	// lookup each JWS key set to validate loading of the respective JWKS key succeeded.
	for _, cfg := range r.issuerToConfig {
		_, err := r.cache.Lookup(ctx, cfg.JWKSURI)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *jwtResolver) Stop(ctx context.Context) error {
	return r.cache.Shutdown(ctx)
}

// ResolveSubject, token is expected to be a string in JWS compact serialization format.
func (r *jwtResolver) ResolveSubject(ctx context.Context, token string) (string, error) {
	unverifiedOptions := []jwt.ParseOption{
		jwt.WithVerify(false),
		jwt.WithValidate(false),
	}

	unverifiedToken, err := jwt.ParseString(token, unverifiedOptions...)
	if err != nil {
		return "", err
	}

	unverifiedIssuer, ok := unverifiedToken.Issuer()
	if !ok {
		return "", err
	}

	// if not OIDC config URLs registered, we can only resolve insecurely as there are no JKWS key sets.
	if len(r.issuerToConfig) == 0 {
		return r.resolveSubjectInsecure(token)
	}

	// look up JKWSURI from OIDC config using the unverified issuer value.
	cfg, ok := r.issuerToConfig[unverifiedIssuer]
	if !ok {
		return "", errors.Errorf("No mapping found of issuer %q to JWKS URI", unverifiedIssuer)
	}

	// fetch JWKS key set from the cache.
	keySet, err := r.cache.Fetch(ctx, cfg.JWKSURI)
	if err != nil {
		return "", err
	}

	verifiedOptions := []jwt.ParseOption{
		// jwt.WithVerify(false),
		jwt.WithValidate(true),
		jwt.WithAcceptableSkew(r.jwtConfig.AcceptableTimeSkewDuration()),
		jwt.WithKeySet(keySet),
	}

	// verify JWT using JWKS key set.
	verifiedToken, err := jwt.ParseString(token, verifiedOptions...)
	if err != nil {
		return "", err
	}

	// get the subject used to resolve the to the User object instance.
	subject, ok := verifiedToken.Subject()
	if !ok {
		return "", errors.Errorf("no verified subject found")
	}

	return subject, nil
}

func (r *jwtResolver) resolveSubjectInsecure(token string) (string, error) {
	verifiedOptions := []jwt.ParseOption{
		jwt.WithVerify(false),
		jwt.WithValidate(true),
		jwt.WithAcceptableSkew(r.jwtConfig.AcceptableTimeSkewDuration()),
	}

	// verify JWT
	verifiedToken, err := jwt.ParseString(token, verifiedOptions...)
	if err != nil {
		return "", err
	}

	// get the subject used to resolve the to the User object instance.
	subject, ok := verifiedToken.Subject()
	if !ok {
		return "", errors.Errorf("no verified subject found")
	}

	return subject, nil
}

// OIDCConfiguration holds the structural properties returned by the OpenID discovery endpoint.
// We only define the subset of fields we use.
type OIDCConfiguration struct {
	Issuer   string   `json:"issuer"`
	JWKSURI  string   `json:"jwks_uri"`
	AuthURI  string   `json:"authorization_endpoint,omitempty"`
	TokenURI string   `json:"token_endpoint,omitempty"`
	Algs     []string `json:"id_token_signing_alg_values_supported,omitempty"`
}

const fetchOIDCConfigReqTimeout = 5 * time.Second

func fetchOIDCConfig(ctx context.Context, discoveryURL string) (*OIDCConfiguration, error) {
	httpClient := &http.Client{
		Timeout: fetchOIDCConfigReqTimeout,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, status.Errorf(codes.Internal, "unexpected status code: %d", resp.StatusCode)
	}

	var config OIDCConfiguration
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config json: %w", err)
	}

	return &config, nil
}
