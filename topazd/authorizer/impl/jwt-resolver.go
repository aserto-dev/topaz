package impl

import (
	"context"

	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/pkg/errors"

	"github.com/jwx-go/jwkfetch/v4"
	"github.com/lestrrat-go/httprc/v3"
	"github.com/lestrrat-go/jwx/v4/jwt"
)

type jwtResolver struct {
	cache          *jwkfetch.Cache
	jwtConfig      *config.JWT
	issuerToConfig map[string]*OidcConfiguration
	oidcClient     *OidcClient
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
		issuerToConfig: make(map[string]*OidcConfiguration),
		oidcClient:     NewOidcClient(),
	}

	return &resolver, nil
}

func (r *jwtResolver) Start(ctx context.Context) error {
	// hydrate the configs map for each allowed issuer, registered in config.jwt.allowed_issuers.
	for _, configAllowedIssuer := range r.jwtConfig.AllowedIssuers {
		config, err := r.oidcClient.FetchAndValidateConfig(ctx, configAllowedIssuer)
		if err != nil {
			return err
		}

		r.issuerToConfig[configAllowedIssuer] = config
	}

	// register each endpoints JWKS URI.
	for _, cfg := range r.issuerToConfig {
		err := r.cache.Register(
			ctx,
			cfg.JwksURI,
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
		_, err := r.cache.Lookup(ctx, cfg.JwksURI)
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
	if !ok || unverifiedIssuer == "" {
		return "", errors.Errorf("unverified token does not contain an issuer field")
	}

	// if not OIDC config URLs registered, we can only resolve insecurely as there are no JWKS key sets.
	if len(r.issuerToConfig) == 0 {
		return r.resolveSubjectInsecure(token)
	}

	// look up JKWS URI from OIDC config using the unverified issuer value.
	cfg, ok := r.issuerToConfig[unverifiedIssuer]
	if !ok {
		return "", errors.Errorf("No mapping found of issuer %q to JWKS URI", unverifiedIssuer)
	}

	// fetch JWKS key set from the cache.
	keySet, err := r.cache.Fetch(ctx, cfg.JwksURI)
	if err != nil {
		return "", err
	}

	verifiedOptions := []jwt.ParseOption{
		jwt.WithValidate(true),
		jwt.WithAcceptableSkew(r.jwtConfig.AcceptableTimeSkewDuration()),
		jwt.WithKeySet(keySet),
	}

	if r.jwtConfig.ExpectedAudience != "" {
		verifiedOptions = append(verifiedOptions, jwt.WithAudience(r.jwtConfig.ExpectedAudience))
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
