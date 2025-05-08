package impl

import (
	"context"
	"encoding/json"
	goruntime "runtime"
	"sync"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/open-policy-agent/opa/v1/server/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	runtime "github.com/aserto-dev/runtime"

	"github.com/aserto-dev/topaz/pkg/version"
	"github.com/aserto-dev/topaz/resolvers"
)

const (
	InputUser     string = "user"
	InputIdentity string = "identity"
	InputPolicy   string = "policy"
	InputResource string = "resource"
)

type AuthorizerServer struct {
	logger      *zerolog.Logger
	issuers     sync.Map
	jwkCache    *jwk.Cache
	jwtTimeSkew time.Duration
	dsClient    dsr3.ReaderClient
	rtResolver  resolvers.RuntimeResolver
}

func NewAuthorizerServer(
	ctx context.Context,
	dsClient dsr3.ReaderClient,
	rtResolver resolvers.RuntimeResolver,
	jwtTimeSkew time.Duration,
) *AuthorizerServer {
	newLogger := zerolog.Ctx(ctx).With().Str("component", "authorizer").Logger()

	jwkCache := jwk.NewCache(ctx)

	return &AuthorizerServer{
		logger:      &newLogger,
		jwkCache:    jwkCache,
		jwtTimeSkew: jwtTimeSkew,
		dsClient:    dsClient,
		rtResolver:  rtResolver,
	}
}

func (s *AuthorizerServer) Info(ctx context.Context, req *authorizer.InfoRequest) (*authorizer.InfoResponse, error) {
	buildVersion := version.GetInfo()

	res := &authorizer.InfoResponse{
		Version: buildVersion.Version,
		Commit:  buildVersion.Commit,
		Date:    buildVersion.Date,
		Os:      goruntime.GOOS,
		Arch:    goruntime.GOARCH,
	}

	return res, nil
}

func (s *AuthorizerServer) getRuntime(ctx context.Context, policyInstance *api.PolicyInstance) (*runtime.Runtime, error) {
	if policyInstance != nil {
		rt, err := s.rtResolver.RuntimeFromContext(ctx, policyInstance.GetName())
		if err != nil {
			return nil, errors.Wrap(err, "failed to procure tenant runtime")
		}

		return rt, err
	}

	rt, err := s.rtResolver.RuntimeFromContext(ctx, "")
	if err != nil {
		return nil, aerr.ErrInvalidPolicyID.Msg("undefined policy context")
	}

	return rt, err
}

func (s *AuthorizerServer) resolveIdentityContext(ctx context.Context, idCtx *api.IdentityContext, input map[string]any) error {
	log := s.logger.With().Str("api", "authz").Logger()

	if idCtx.GetType() != api.IdentityType_IDENTITY_TYPE_NONE {
		input[InputIdentity] = convert(idCtx)

		user, err := s.getUserFromIdentityContext(ctx, idCtx)
		if err != nil || user == nil {
			log.Error().Err(err).Interface("identity_context", idCtx).Msg("failed to resolve identity context")

			return aerr.ErrAuthenticationFailed.WithGRPCStatus(codes.NotFound).Msg("failed to resolve identity context")
		}

		input[InputUser] = convert(user)
	}

	return nil
}

func traceLevelToExplainModeV2(t authorizer.TraceLevel) types.ExplainModeV1 {
	switch t {
	case authorizer.TraceLevel_TRACE_LEVEL_UNKNOWN:
		return types.ExplainOffV1
	case authorizer.TraceLevel_TRACE_LEVEL_OFF:
		return types.ExplainOffV1
	case authorizer.TraceLevel_TRACE_LEVEL_FULL:
		return types.ExplainFullV1
	case authorizer.TraceLevel_TRACE_LEVEL_NOTES:
		return types.ExplainNotesV1
	case authorizer.TraceLevel_TRACE_LEVEL_FAILS:
		return types.ExplainFailsV1
	default:
		return types.ExplainOffV1
	}
}

// convert, explicitly converts from proto message any in order
// to preserve enum values as strings when marshaled to JSON.
func convert(msg proto.Message) any {
	b, err := protojson.MarshalOptions{
		Multiline:       false,
		Indent:          "  ",
		AllowPartial:    false,
		UseProtoNames:   true,
		UseEnumNumbers:  false,
		EmitUnpopulated: true,
	}.Marshal(msg)
	if err != nil {
		return nil
	}

	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return nil
	}

	return v
}
