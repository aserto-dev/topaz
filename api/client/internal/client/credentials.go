package client

import (
	"context"
	"strings"
)

// TokenAuth bearer token based authentication.
//
// It implements the interface credentials.PerRPCCredentials.
type TokenAuth struct {
	token string
}

func NewTokenAuth(token string) *TokenAuth {
	pieces := strings.Split(token, " ")
	if len(pieces) == 1 {
		return &TokenAuth{
			token: Bearer + " " + token,
		}
	}

	return &TokenAuth{
		token: token,
	}
}

func (t TokenAuth) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	return map[string]string{
		Authorization: t.token,
	}, nil
}

func (TokenAuth) RequireTransportSecurity() bool {
	return false
}

// APIKeyAuth API key based authentication.
//
// It implements the interface credentials.PerRPCCredentials.
type APIKeyAuth struct {
	key string
}

func NewAPIKeyAuth(key string) *APIKeyAuth {
	return &APIKeyAuth{
		key: key,
	}
}

func (k *APIKeyAuth) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	return map[string]string{
		Authorization: Basic + " " + k.key,
	}, nil
}

func (k *APIKeyAuth) RequireTransportSecurity() bool {
	return false
}
