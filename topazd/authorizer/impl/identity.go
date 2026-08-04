package impl

import (
	"context"
	"time"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	dsc "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsr "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/go-directory/pkg/pb"
	"github.com/aserto-dev/topaz/pkg/grpcc"
	"github.com/aserto-dev/topaz/topazd/directory"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
)

const identityResolutionTimeout = 60 * time.Second

// resolveIdentityContext.
func (s *AuthorizerServer) resolveIdentityContext(ctx context.Context, identityContext *api.IdentityContext, input map[string]any) error {
	if identityContext == nil {
		return aerr.ErrInvalidArgument.Msg("identity context not set")
	}

	// nothing to resolve when identity type equals IdentityType_IDENTITY_TYPE_NONE.
	if identityContext.GetType() == api.IdentityType_IDENTITY_TYPE_NONE {
		return nil
	}

	// context control timeout for end-to-end identity (JWT when used) and directory identity to  user resolution.
	ctx, cancel := context.WithTimeout(ctx, identityResolutionTimeout)
	defer cancel()

	// Step 1: save identity context in to input.identity.
	input[InputIdentity] = grpcc.ProtoToAny(identityContext)

	// Step 2: resolve identity from identity context
	identity, err := s.resolveSubjectFromIdentityContext(ctx, identityContext)
	if err != nil {
		return aerr.ErrAuthenticationFailed.WithGRPCStatus(codes.NotFound).Msg("failed to resolve identity context")
	}

	// if IDENTITY_TYPE_MANUAL, there resulting user object is an empty JSON object.
	if identityContext.GetType() == api.IdentityType_IDENTITY_TYPE_MANUAL {
		input[InputUser] = pb.NewStruct()
		return nil
	}

	// Step 3: resolve user from identity.
	user, err := s.resolveUserFromSubject(ctx, identity)
	if err != nil {
		return aerr.ErrAuthenticationFailed.WithGRPCStatus(codes.NotFound).Msg("failed to resolve user from identity")
	}

	// Step 4: add user object to input.user.
	input[InputUser] = grpcc.ProtoToAny(user)

	return nil
}

func (s *AuthorizerServer) resolveSubjectFromIdentityContext(ctx context.Context, identityContext *api.IdentityContext) (string, error) {
	if identityContext.GetIdentity() == "" {
		return "", errors.Errorf("identity value not set (type: %s)", identityContext.GetType().String())
	}

	switch identityContext.GetType() {
	case api.IdentityType_IDENTITY_TYPE_SUB:
		return identityContext.GetIdentity(), nil

	case api.IdentityType_IDENTITY_TYPE_JWT:
		return s.jwtResolver.ResolveSubject(ctx, identityContext.GetIdentity())

	case api.IdentityType_IDENTITY_TYPE_MANUAL:
		return identityContext.GetIdentity(), nil

	case api.IdentityType_IDENTITY_TYPE_NONE:
		fallthrough

	default:
		return "", errors.Errorf("invalid identity type %s", identityContext.GetType().String())
	}
}

func (s *AuthorizerServer) resolveUserFromSubject(ctx context.Context, subject string) (*dsc.Object, error) {
	client := dsr.NewReaderClient(s.resolver.GetDirectoryResolver().GetConn())

	objResp, err := directory.ResolveIdentity(ctx, client, subject)
	if err != nil {
		return nil, err
	}

	return objResp, nil
}
