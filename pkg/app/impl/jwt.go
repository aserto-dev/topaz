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
	v2 "github.com/aserto-dev/go-directory/aserto/directory/common/v2"
	ds2 "github.com/aserto-dev/go-directory/aserto/directory/v2"
	"github.com/aserto-dev/topaz/builtins/edge/ds"
	"github.com/aserto-dev/topaz/pkg/app/instance"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

var (
	// ErrMissingMetadata - metadata element missing
	ErrMissingMetadata = aerr.ErrInvalidArgument.Msg("missing metadata")
	// ErrMissingToken - token missing from metadata
	ErrMissingToken = aerr.ErrInvalidArgument.Msg("missing token")
	// ErrInvalidToken - token not valid
	ErrInvalidToken = aerr.ErrAuthenticationFailed.Msg("invalid token")
)

// getUserFromJWT
func (s *AuthorizerServer) getUserFromJWT(ctx context.Context, bearerJWT string) (proto.Message, error) {
	resp := v2.Object{}

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

// getIdentityFromJWT
func (s *AuthorizerServer) getIdentityFromJWT(ctx context.Context, bearerJWT string) (string, error) {
	log := instance.GetInstanceLogger(ctx, s.logger)

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

// jwksURL
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

// getUserFromIdentityContext .
func (s *AuthorizerServer) getUserFromIdentityContext(ctx context.Context, identityContext *api.IdentityContext) (proto.Message, error) {
	if identityContext == nil {
		return nil, aerr.ErrInvalidArgument.Msg("identity context not set")
	}

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
	return s.getUserFromIdentityV2(ctx, identity)
}

// func (s *AuthorizerServer) getUserFromIdentityV1(ctx context.Context, identity string) (proto.Message, error) {
// 	tenantID := instance.ExtractID(ctx)

// 	directory, err := s.directoryResolver.DirectoryFromContext(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	userV1, err := directory.GetUserFromIdentity(tenantID, identity)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return userV1, nil
// }

func (s *AuthorizerServer) getUserFromIdentityV2(ctx context.Context, identity string) (proto.Message, error) {
	uid, err := s.getIdentityV2(ctx, identity)
	switch {
	case errors.Is(err, aerr.ErrDirectoryObjectNotFound):
		if !ds.IsValidID(identity) {
			return nil, err
		}
		uid = identity

	case err != nil:
		return nil, err

	default:
	}

	user, err := s.getUserV2(ctx, uid)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthorizerServer) getIdentityV2(ctx context.Context, identity string) (string, error) {
	client, err := s.directoryResolver.GetDS(ctx)
	if err != nil {
		return "", err
	}
	identityString := "identity"
	identityResp, err := client.GetObject(ctx, &ds2.GetObjectRequest{
		Param: &v2.ObjectIdentifier{
			Type: &identityString,
			Key:  &identity,
		},
	})
	if err != nil {
		return "", err
	}

	if identityResp.Result == nil {
		return "", aerr.ErrDirectoryObjectNotFound
	}

	iid := identityResp.Result.Id
	relationString := "identifier"
	subjectType := "user"

	relResp, err := client.GetRelation(ctx, &ds2.GetRelationRequest{
		Param: &v2.RelationIdentifier{
			Object:   &v2.ObjectIdentifier{Type: &identityString, Id: &iid},
			Relation: &v2.RelationTypeIdentifier{Name: &relationString, ObjectType: &identityString},
			Subject:  &v2.ObjectIdentifier{Type: &subjectType},
		},
	})
	if err != nil {
		return "", err
	}

	if relResp.Result == nil {
		return "", aerr.ErrDirectoryObjectNotFound
	}

	uid := relResp.Result.Subject.Id

	return *uid, nil
}

func (s *AuthorizerServer) getUserV2(ctx context.Context, uid string) (*v2.Object, error) {
	client, err := s.directoryResolver.GetDS(ctx)
	if err != nil {
		return nil, err
	}

	userResp, err := client.GetObject(ctx, &ds2.GetObjectRequest{
		Param: &v2.ObjectIdentifier{
			Id: &uid,
		},
	})
	if err != nil {
		return nil, err
	}

	if userResp.Result == nil {
		return nil, aerr.ErrDirectoryObjectNotFound
	}

	return userResp.Result, nil
}
