package impl

import (
	"context"
	"encoding/json"
	goruntime "runtime"
	"sync"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	runtime "github.com/aserto-dev/runtime"

	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/topazd/authorizer/resolvers"
	"github.com/aserto-dev/topaz/topazd/version"

	"github.com/jwx-go/jwkfetch/v4"
	"github.com/lestrrat-go/httprc/v3"
	"github.com/open-policy-agent/opa/v1/server/types"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	InputUser     string = "user"
	InputIdentity string = "identity"
	InputPolicy   string = "policy"
	InputResource string = "resource"
)

type AuthorizerServer struct {
	cfg      *config.Common
	logger   *zerolog.Logger
	issuers  sync.Map
	jwkCache *jwkfetch.Cache

	resolver *resolvers.Resolvers

	// preparedQueries memoizes the rego.PreparedEvalQuery values produced
	// for each (policy path, decisions) tuple seen via Is(). Without this
	// cache every Is() call re-parses + re-plans the same Rego query and
	// serializes goroutines on the OPA compiler's internal structures.
	preparedQueries *preparedQueryCache
}

func NewAuthorizerServer(
	ctx context.Context,
	logger *zerolog.Logger,
	cfg *config.Common,
	rf *resolvers.Resolvers,
) (*AuthorizerServer, error) {
	newLogger := logger.With().Str("component", "api.grpc").Logger()

	jwkCache, err := jwkfetch.NewCache(ctx, httprc.NewClient())
	if err != nil {
		return nil, err
	}

	go func() { //nolint:gosec // G118 - cleanup cannot use request context as it is already cancelled.
		<-ctx.Done()

		cleanupCtx := context.Background()

		_ = jwkCache.Shutdown(cleanupCtx)
	}()

	return &AuthorizerServer{
		cfg:             cfg,
		logger:          &newLogger,
		resolver:        rf,
		jwkCache:        jwkCache,
		preparedQueries: newPreparedQueryCache(),
	}, nil
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

func (s *AuthorizerServer) getRuntime(ctx context.Context) (*runtime.Runtime, error) {
	rt, err := s.resolver.GetRuntimeResolver().GetRuntime(ctx)
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
		Multiline:         false,
		Indent:            "  ",
		AllowPartial:      false,
		UseProtoNames:     true,
		UseEnumNumbers:    false,
		EmitUnpopulated:   true,
		EmitDefaultValues: false,
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
