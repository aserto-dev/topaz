package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	dsc "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsr "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/go-directory/pkg/pb"
	"github.com/aserto-dev/topaz/topazd/directory"

	"github.com/jwx-go/jwkfetch/v4"
	"github.com/lestrrat-go/jwx/v4/jwt"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const jwtMinRefreshInterval = 15 * time.Minute

var (
	// ErrMissingMetadata - metadata element missing.
	ErrMissingMetadata = aerr.ErrInvalidArgument.Msg("missing metadata")
	// ErrMissingToken - token missing from metadata.
	ErrMissingToken = aerr.ErrInvalidArgument.Msg("missing token")
	// ErrInvalidToken - token not valid.
	ErrInvalidToken = aerr.ErrAuthenticationFailed.Msg("invalid token")
)

// getUserFromJWT.
func (s *AuthorizerServer) getUserFromJWT(ctx context.Context, bearerJWT string) (proto.Message, error) {
	resp := dsc.Object{}

	ident, err := s.getIdentityFromJWT(ctx, bearerJWT)
	if err != nil {
		return &resp, err
	}

	user, err := s.getUserFromIdentity(ctx, ident)
	if err != nil {
		return &resp, err
	}

	return user, nil
}

// getIdentityFromJWT.
func (s *AuthorizerServer) getIdentityFromJWT(ctx context.Context, bearerJWT string) (string, error) {
	log := s.logger

	jwtTemp, err := jwt.ParseString(bearerJWT, jwt.WithVerify(false))
	if err != nil {
		log.Error().Err(err).Msg("jwt parse without validation")
		return "", err
	}

	options, err := s.jwtParseStringOptions(ctx, jwtTemp)
	if err != nil {
		return "", err
	}

	jwtToken, err := jwt.ParseString(
		bearerJWT,
		options...,
	)
	if err != nil {
		log.Error().Err(err).Msg("jwt parse with validation")
		return "", err
	}

	ident, ok := jwtToken.Subject()
	if !ok || ident == "" {
		return "", errors.Errorf("no sub field present in token")
	}

	return ident, nil
}

func (s *AuthorizerServer) jwtParseStringOptions(ctx context.Context, jwtToken jwt.Token) ([]jwt.ParseOption, error) {
	options := []jwt.ParseOption{
		jwt.WithValidate(true),
		jwt.WithAcceptableSkew(time.Duration(s.cfg.JWT.AcceptableTimeSkewSeconds) * time.Second),
	}

	issuer, ok := jwtToken.Issuer()
	if !ok {
		return nil, errors.Errorf("no iss field present in token")
	}

	// NOTE: if config.jwt.allowed_issuers is empty, there is no enforcement of the issuer URL.
	if len(s.cfg.JWT.AllowedIssuers) > 0 && !issuerAllowed(s.cfg.JWT.AllowedIssuers, issuer) {
		return nil, errors.Errorf("issuer %q is not an allowed issuer, see: config.jwt.allowed_issuers", issuer)
	}

	jwtKeysURL, err := s.jwksURLFromCache(ctx, issuer)
	if err != nil {
		return nil, errors.Wrap(err, "token didn't have a JWKS endpoint we could use for verification")
	} else {
		err := registerJWKSURL(ctx, s.jwkCache, jwtKeysURL)
		if err != nil {
			return nil, errors.Wrap(err, "failed to register JWKS URL")
		}

		jwkSet, errX := s.jwkCache.Lookup(ctx, jwtKeysURL)
		if errX != nil {
			return nil, errors.Wrap(errX, "failed to fetch JWK set for validation")
		}

		options = append(options, jwt.WithKeySet(jwkSet))
	}

	return options, nil
}

// issuerAllowed reports whether issuer matches one of the configured allowed
// issuers. A configured entry matches either exactly, or as a path prefix
// when the entry ends in "/" (e.g. "https://idp.example.com/" matches
// "https://idp.example.com/tenant-a"). Prefix matching is always anchored at
// a "/" boundary so an entry can't be defeated by an attacker-chosen suffix
// such as "https://idp.example.com.attacker.evil".
func issuerAllowed(allowed []string, issuer string) bool {
	return slices.ContainsFunc(allowed, func(cfgISS string) bool {
		if issuer == cfgISS {
			return true
		}

		return strings.HasSuffix(cfgISS, "/") && strings.HasPrefix(issuer, cfgISS)
	})
}

func registerJWKSURL(ctx context.Context, jwkCache *jwkfetch.Cache, jwksURL string) error {
	if !jwkCache.IsRegistered(ctx, jwksURL) {
		err := jwkCache.Register(ctx, jwksURL, jwkfetch.WithMinInterval(jwtMinRefreshInterval))
		if err != nil {
			return err
		}

		if _, err := jwkCache.Refresh(ctx, jwksURL); err != nil {
			fmt.Printf("failed to refresh JWKS: %s\n", err)
			return err
		}
	}

	return nil
}

func (s *AuthorizerServer) jwksURLFromCache(ctx context.Context, issuer string) (string, error) {
	if val, ok := s.issuers.Load(issuer); ok {
		jwksURL, _ := val.(string)
		return jwksURL, nil
	}

	jk, err := s.jwksURL(ctx, issuer)
	if err != nil {
		return "", err
	}

	jwksURL := jk.String()

	s.issuers.Store(issuer, jwksURL)

	return jwksURL, nil
}

func (s *AuthorizerServer) jwksURL(ctx context.Context, baseURL string) (*url.URL, error) {
	const (
		wellknownConfig = `.well-known/openid-configuration`
		wellknownJWKS   = `.well-known/jwks.json`
	)

	b, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	if b.Scheme == "" {
		return nil, errors.New("empty baseURL scheme, must be https or http")
	}

	originalPath := b.Path
	b.Path = filepath.Join(originalPath, wellknownConfig)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.String(), http.NoBody)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err == nil {
		defer func() { _ = resp.Body.Close() }()

		var config struct {
			URI string `json:"jwks_uri"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&config); err == nil {
			if u, err := url.Parse(config.URI); err == nil {
				return u, nil
			}
		}
	}

	// No usable OIDC discovery document; fall back to the
	// plain JWKS well-known path.
	b.Path = filepath.Join(originalPath, wellknownJWKS)

	return b, nil
}

func (s *AuthorizerServer) getUserFromIdentityContext(ctx context.Context, identityContext *api.IdentityContext) (proto.Message, error) {
	if identityContext == nil {
		return nil, aerr.ErrInvalidArgument.Msg("identity context not set")
	}

	switch identityContext.GetType() {
	case api.IdentityType_IDENTITY_TYPE_NONE:
		return nil, nil

	case api.IdentityType_IDENTITY_TYPE_SUB:
		if identityContext.GetIdentity() == "" {
			return nil, errors.Errorf("identity value not set (type: %s)", identityContext.GetType().String())
		}

		user, err := s.getUserFromIdentity(ctx, identityContext.GetIdentity())
		if err != nil {
			return nil, err
		}

		return user, nil
	case api.IdentityType_IDENTITY_TYPE_JWT:
		if identityContext.GetIdentity() == "" {
			return nil, errors.Errorf("identity value not set (type: %s)", identityContext.GetType().String())
		}

		user, err := s.getUserFromJWT(ctx, identityContext.GetIdentity())
		if err != nil {
			return nil, err
		}

		return user, nil
	case api.IdentityType_IDENTITY_TYPE_MANUAL:
		if identityContext.GetIdentity() == "" {
			return nil, errors.Errorf("identity value not set (type: %s)", identityContext.GetType().String())
		}

		// the resulting user object will be an empty object.
		return pb.NewStruct(), nil
	default:
		return nil, errors.Errorf("invalid identity type %s", identityContext.GetType().String())
	}
}

func (s *AuthorizerServer) getUserFromIdentity(ctx context.Context, identity string) (proto.Message, error) {
	client := dsr.NewReaderClient(s.resolver.GetDirectoryResolver().GetConn())

	user, err := directory.ResolveIdentity(ctx, client, identity)

	switch {
	case status.Code(err) == codes.NotFound:
		return s.getUserObject(ctx, identity)
	case err != nil:
		return nil, err
	default:
		return user, nil
	}
}

// getUserObject, retrieves an user object, using the identity as the object_id (legacy).
func (s *AuthorizerServer) getUserObject(ctx context.Context, objID string) (proto.Message, error) {
	client := dsr.NewReaderClient(s.resolver.GetDirectoryResolver().GetConn())

	objResp, err := client.GetObject(ctx, &dsr.GetObjectRequest{
		ObjectType: directory.User,
		ObjectId:   objID,
	})
	if err != nil {
		return nil, err
	}

	return objResp.GetResult(), nil
}
