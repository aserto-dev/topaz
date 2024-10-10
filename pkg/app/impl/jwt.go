package impl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	dsc3 "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/go-directory/pkg/pb"
	"github.com/aserto-dev/topaz/directory"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
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
	resp := dsc3.Object{}

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

	ident := jwtToken.Subject()

	return ident, nil
}

func (s *AuthorizerServer) jwtParseStringOptions(ctx context.Context, jwtToken jwt.Token) ([]jwt.ParseOption, error) {
	options := []jwt.ParseOption{
		jwt.WithValidate(true),
		jwt.WithAcceptableSkew(time.Duration(s.cfg.JWT.AcceptableTimeSkewSeconds) * time.Second),
	}

	jwtKeysURL, err := s.jwksURLFromCache(ctx, jwtToken.Issuer())

	if err != nil {
		return nil, errors.Wrap(err, "token didn't have a JWKS endpoint we could use for verification")
	} else {
		err := registerJWKSURL(ctx, s.jwkCache, jwtKeysURL)
		if err != nil {
			return nil, errors.Wrap(err, "failed to register JWKS URL")
		}

		jwkSet, errX := s.jwkCache.Get(ctx, jwtKeysURL)
		if errX != nil {
			return nil, errors.Wrap(errX, "failed to fetch JWK set for validation")
		}
		options = append(options, jwt.WithKeySet(jwkSet))
	}

	return options, nil
}

func registerJWKSURL(ctx context.Context, jwkCache *jwk.Cache, jwksURL string) error {
	if !jwkCache.IsRegistered(jwksURL) {
		err := jwkCache.Register(jwksURL, jwk.WithMinRefreshInterval(15*time.Minute))
		if err != nil {
			return err
		}

		_, err = jwkCache.Refresh(ctx, jwksURL)
		if err != nil {
			fmt.Printf("failed to refresh JWKS: %s\n", err)
			return err
		}
	}

	return nil
}

func (s *AuthorizerServer) jwksURLFromCache(ctx context.Context, issuer string) (string, error) {
	var jwksURL string
	if val, ok := s.issuers.Load(issuer); ok {
		jwksURL = val.(string)
	} else {
		jk, err := s.jwksURL(ctx, issuer)
		if err != nil {
			return "", err
		}
		jwksURL = jk.String()
		s.issuers.Store(issuer, jwksURL)
	}

	return jwksURL, nil
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
			return nil, errors.Errorf("identity value not set (type: %s)", identityContext.Type.String())
		}

		user, err := s.getUserFromIdentity(ctx, identityContext.Identity)
		if err != nil {
			return nil, err
		}

		return user, nil
	case api.IdentityType_IDENTITY_TYPE_JWT:
		if identityContext.Identity == "" {
			return nil, errors.Errorf("identity value not set (type: %s)", identityContext.Type.String())
		}

		user, err := s.getUserFromJWT(ctx, identityContext.Identity)
		if err != nil {
			return nil, err
		}

		return user, nil
	case api.IdentityType_IDENTITY_TYPE_MANUAL:
		if identityContext.Identity == "" {
			return nil, errors.Errorf("identity value not set (type: %s)", identityContext.Type.String())
		}

		// the resulting user object will be an empty object.
		return pb.NewStruct(), nil
	default:
		return nil, errors.Errorf("invalid identity type %s", identityContext.Type.String())
	}
}

func (s *AuthorizerServer) getUserFromIdentity(ctx context.Context, identity string) (proto.Message, error) {
	client := s.resolver.GetDirectoryResolver().GetDS()

	user, err := directory.GetIdentityV2(ctx, client, identity)
	switch {
	case errors.Is(err, aerr.ErrDirectoryObjectNotFound):
		// Try to find a user with key == identity
		return s.getObject(ctx, "user", identity)
	case err != nil:
		return nil, err
	default:
		return addObjectKey(user)
	}
}

func (s *AuthorizerServer) getObject(ctx context.Context, objType, objID string) (proto.Message, error) {
	client := s.resolver.GetDirectoryResolver().GetDS()

	objResp, err := client.GetObject(ctx, &dsr3.GetObjectRequest{
		ObjectType: objType,
		ObjectId:   objID,
	})
	if err != nil {
		return nil, err
	}

	return addObjectKey(objResp.Result)
}

func addObjectKey(in *dsc3.Object) (proto.Message, error) {
	buf := new(bytes.Buffer)
	if err := pb.ProtoToBuf(buf, in); err != nil {
		return nil, err
	}

	out := pb.NewStruct()

	if err := pb.BufToProto(bytes.NewReader(buf.Bytes()), out); err != nil {
		return nil, err
	}

	out.Fields["key"] = structpb.NewStringValue(out.Fields["id"].GetStringValue())

	return out, nil
}
