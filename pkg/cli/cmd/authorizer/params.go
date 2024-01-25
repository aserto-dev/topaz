package authorizer

import (
	"encoding/json"
	"errors"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"google.golang.org/protobuf/types/known/structpb"
)

type IdentityType string

var ErrResourceNotJSON = errors.New("resource is not a JSON object")

const (
	IdentityTypeNone IdentityType = "none"
	IdentityTypeSub  IdentityType = "sub"
	IdentityTypeJwt  IdentityType = "jwt"
)

type AuthParams struct {
	Identity     string       `name:"identity" help:"caller identity" default:""`
	IdentityType IdentityType `name:"identity-type" enum:"sub,jwt,none" help:"type of identity [sub|jwt|none]"  default:"none"`
	Resource     string       `name:"resource" help:"a JSON object to include as resource context"`
	PolicyID     string       `name:"policy-id" required:"" help:"policy id"`
}

func (a AuthParams) IdentityContext() *api.IdentityContext {
	var idType api.IdentityType
	switch a.IdentityType {
	case IdentityTypeSub:
		idType = api.IdentityType_IDENTITY_TYPE_SUB
	case IdentityTypeJwt:
		idType = api.IdentityType_IDENTITY_TYPE_JWT
	case IdentityTypeNone:
		idType = api.IdentityType_IDENTITY_TYPE_NONE
	}

	return &api.IdentityContext{
		Identity: a.Identity,
		Type:     idType,
	}
}

func (a AuthParams) ResourceContext() (*structpb.Struct, error) {
	result := &structpb.Struct{}

	if a.Resource != "" {
		var r interface{}
		if err := json.Unmarshal([]byte(a.Resource), &r); err != nil {
			return result, err
		}

		m, ok := r.(map[string]interface{})
		if !ok {
			return result, ErrResourceNotJSON
		}

		return structpb.NewStruct(m)
	}

	return result, nil
}
