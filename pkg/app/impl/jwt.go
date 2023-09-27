package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	dsc2 "github.com/aserto-dev/go-directory/aserto/directory/common/v2"
	dsr2 "github.com/aserto-dev/go-directory/aserto/directory/reader/v2"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/aserto-dev/topaz/directory"
)

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
	resp := dsc2.Object{}

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

	jwtTemp, err := jwt.ParseString(bearerJWT, jwt.WithValidate(false))
	if err != nil {
		log.Error().Err(err).Msg("jwt parse without validation")
		return "", err
	}

	options := []jwt.ParseOption{
		jwt.WithValidate(true),
		jwt.WithAcceptableSkew(time.Duration(s.cfg.JWT.AcceptableTimeSkewSeconds) * time.Second),
	}

	jwksURL, err := s.jwksURL(ctx, jwtTemp.Issuer())
	if err != nil {
		log.Debug().Str("issuer", jwtTemp.Issuer()).Msg("token didn't have a JWKS endpoint we could use for verification")
	} else {
		jwkSet, errX := jwk.Fetch(ctx, jwksURL.String())
		if errX != nil {
			return "", errors.Wrap(errX, "failed to fetch JWK set for validation")
		}

		options = append(options, jwt.WithKeySet(jwkSet))
	}

	jwtToken, err := jwt.ParseString(
		bearerJWT,
		options...,
	)
	if err != nil {
		log.Error().Err(err).Msg("jwt parse with validation")
		return "", err
	}

	ident := jwtToken.Subject()

	return ident, nil
}

// jwksURL.
func (s *AuthorizerServer) jwksURL(ctx context.Context, baseURL string) (*url.URL, error) {
	const (
		wellknownConfig = `.well-known/openid-configuration`
		wellknownJWKS   = `.well-known/jwks.json`
	)

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "" {
		return nil, errors.New("no scheme defined for baseURL")
	}

	originalPath := u.Path
	u.Path = path.Join(originalPath, wellknownConfig)

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), http.NoBody)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		var config struct {
			URI string `json:"jwks_uri"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&config); err == nil {
			if u, err = url.Parse(config.URI); err == nil {
				return u, nil
			}
		}
	}

	u.Path = path.Join(originalPath, wellknownJWKS)

	return u, nil
}

// getUserFromIdentityContext.
func (s *AuthorizerServer) getUserFromIdentityContext(ctx context.Context, identityContext *api.IdentityContext) (proto.Message, error) {
	if identityContext == nil {
		return nil, aerr.ErrInvalidArgument.Msg("identity context not set")
	}

	// nolint: exhaustive
	switch identityContext.Type {
	case api.IdentityType_IDENTITY_TYPE_NONE:
		return nil, nil

	case api.IdentityType_IDENTITY_TYPE_SUB:
		if identityContext.Identity == "" {
			return nil, fmt.Errorf("identity value not set (type: %s)", identityContext.Type.String())
		}

		user, err := s.getUserFromIdentity(ctx, identityContext.Identity)
		if err != nil {
			return nil, err
		}

		return user, nil
	case api.IdentityType_IDENTITY_TYPE_JWT:
		if identityContext.Identity == "" {
			return nil, fmt.Errorf("identity value not set (type: %s)", identityContext.Type.String())
		}

		user, err := s.getUserFromJWT(ctx, identityContext.Identity)
		if err != nil {
			return nil, err
		}

		return user, nil
	default:
		return nil, fmt.Errorf("invalid identity type %s", identityContext.Type.String())
	}
}

func (s *AuthorizerServer) getUserFromIdentity(ctx context.Context, identity string) (proto.Message, error) {
	client, err := s.resolver.GetDirectoryResolver().GetDS(ctx)
	if err != nil {
		return nil, err
	}
	user, err := directory.GetIdentityV2(client, ctx, identity)
	switch {
	case errors.Is(err, aerr.ErrDirectoryObjectNotFound):
		// Try to find a user with key == identity
		return s.getObject(ctx, "user", identity)
	case err != nil:
		return nil, err
	default:
		return user, nil
	}
}

func (s *AuthorizerServer) getObject(ctx context.Context, objType, key string) (proto.Message, error) {
	client, err := s.resolver.GetDirectoryResolver().GetDS(ctx)
	if err != nil {
		return nil, err
	}

	objResp, err := client.GetObject(ctx, &dsr2.GetObjectRequest{Param: &dsc2.ObjectIdentifier{Type: &objType, Key: &key}})
	if err != nil {
		return nil, err
	}

	return objResp.Result, nil
}
