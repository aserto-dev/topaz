package impl

import (
	"context"
	goruntime "runtime"
	"time"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	"github.com/aserto-dev/topaz/internal/runtime"

	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/topazd/authorizer/resolvers"
	"github.com/aserto-dev/topaz/topazd/version"

	"github.com/open-policy-agent/opa/v1/server/types"
	"github.com/rs/zerolog"
)

const (
	InputUser     string = "user"
	InputIdentity string = "identity"
	InputPolicy   string = "policy"
	InputResource string = "resource"
)

const cleanupTimeout = 30 * time.Second

type AuthorizerServer struct {
	cfg         *config.Common
	logger      *zerolog.Logger
	jwtResolver *jwtResolver
	resolver    *resolvers.Resolvers

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

	jwtResolver, err := NewJWTResolver(ctx, &cfg.JWT)
	if err != nil {
		return nil, err
	}

	if err := jwtResolver.Start(ctx); err != nil {
		return nil, err
	}

	go func() { //nolint:gosec // G118 - cleanup cannot use request context as it is already cancelled.
		<-ctx.Done()

		cleanupCtx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
		defer cancel()

		_ = jwtResolver.Stop(cleanupCtx)
	}()

	return &AuthorizerServer{
		cfg:             cfg,
		logger:          &newLogger,
		jwtResolver:     jwtResolver,
		resolver:        rf,
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
