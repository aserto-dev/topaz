package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aserto-dev/aserto-grpc/grpcutil"
	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	api "github.com/aserto-dev/go-grpc/aserto/api/v1"
	dl "github.com/aserto-dev/go-grpc/aserto/decision_logs/v1"
	"github.com/aserto-dev/go-lib/ids"
	"github.com/aserto-dev/go-utils/cerr"
	"github.com/aserto-dev/go-utils/pb"
	"github.com/aserto-dev/go-utils/protoutil"
	runtime "github.com/aserto-dev/runtime"
	decisionlog_plugin "github.com/aserto-dev/topaz/decision_log/plugin"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/google/uuid"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/server/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	InputUser     string = "user"
	InputIdentity string = "identity"
	InputPolicy   string = "policy"
	InputResource string = "resource"
)

type AuthorizerServer struct {
	cfg    *config.Common
	logger *zerolog.Logger

	runtimeResolver   resolvers.RuntimeResolver
	directoryResolver resolvers.DirectoryResolver
}

func NewAuthorizerServer(
	logger *zerolog.Logger,
	cfg *config.Common,
	runtimeResolver resolvers.RuntimeResolver,
	directoryResolver resolvers.DirectoryResolver) *AuthorizerServer {

	newLogger := logger.With().Str("component", "api.grpc").Logger()
	return &AuthorizerServer{
		cfg:               cfg,
		runtimeResolver:   runtimeResolver,
		logger:            &newLogger,
		directoryResolver: directoryResolver,
	}
}

func (s *AuthorizerServer) DecisionTree(ctx context.Context, req *authz.DecisionTreeRequest) (*authz.DecisionTreeResponse, error) { // nolint:funlen,gocyclo //TODO: split into smaller functions after merge with onebox
	log := grpcutil.CompleteLogger(ctx, s.logger)

	resp := &authz.DecisionTreeResponse{}

	if req.PolicyContext == nil {
		return resp, cerr.ErrInvalidArgument.Msg("policy context not set")
	}

	if req.PolicyContext.GetId() == "" {
		return resp, cerr.ErrInvalidArgument.Msg("policy context id not set")
	}

	if req.ResourceContext == nil {
		req.ResourceContext = pb.NewStruct()
	}

	if req.Options == nil {
		req.Options = &authz.DecisionTreeOptions{
			PathSeparator: authz.PathSeparator_PATH_SEPARATOR_DOT,
		}
	}

	if req.IdentityContext == nil {
		return resp, cerr.ErrInvalidArgument.Msg("identity context not set")
	}

	if req.IdentityContext.Type == api.IdentityType_IDENTITY_TYPE_UNKNOWN {
		return resp, cerr.ErrInvalidArgument.Msg("identity type UNKNOWN")
	}

	user, err := s.getUserFromIdentityContext(ctx, req.IdentityContext)
	if err != nil {
		log.Error().Err(err).Interface("req", req).Msg("failed to resolve identity context")
		return resp, cerr.ErrAuthenticationFailed.WithGRPCStatus(codes.NotFound).Msg("failed to resolve identity context")
	}

	input := map[string]interface{}{
		InputUser:     convert(user),
		InputIdentity: convert(req.IdentityContext),
		InputPolicy:   req.PolicyContext,
		InputResource: req.ResourceContext,
	}

	policyRuntime, err := s.runtimeResolver.RuntimeFromContext(ctx, req.PolicyContext.GetId(), req.PolicyContext.GetName(), req.PolicyContext.InstanceLabel)
	if err != nil {
		return resp, errors.Wrap(err, "failed to procure tenant runtime")
	}

	policyList, err := policyRuntime.GetPolicyList(
		ctx,
		req.PolicyContext.GetId(),
		policyRuntime.PathFilter(req.Options.PathSeparator, req.PolicyContext.Path),
	)
	if err != nil {
		return resp, errors.Wrap(err, "get policy list")
	}

	decisionFilter := initDecisionFilter(req.PolicyContext.Decisions)

	results := make(map[string]interface{})

	policyContext := &api.PolicyContext{}
	proto.Merge(policyContext, req.PolicyContext)

	for _, policy := range policyList {
		queryStmt := "x = data." + policy.PackageName

		policyContext.Path = policy.PackageName
		input[InputPolicy] = policyContext

		qry, err := rego.New(
			rego.Compiler(policyRuntime.GetPluginsManager().GetCompiler()),
			rego.Store(policyRuntime.GetPluginsManager().Store),
			rego.Input(input),
			rego.Query(queryStmt),
		).PrepareForEval(ctx)

		if err != nil {
			return resp, cerr.ErrBadQuery.Err(err).Str("query", queryStmt)
		}

		packageName := policy.Package(req.Options.PathSeparator)

		queryResults, err := qry.Eval(ctx, rego.EvalInput(input))

		if err != nil {
			return resp, cerr.ErrBadQuery.Err(err).Str("query", queryStmt).Msg("query evaluation failed")
		} else if len(queryResults) == 0 {
			return resp, cerr.ErrBadQuery.Err(err).Str("query", queryStmt).Msg("undefined results")
		}

		if result, ok := queryResults[0].Bindings["x"].(map[string]interface{}); ok {
			for k, v := range result {
				if !decisionFilter(k) {
					continue
				}
				if _, ok := v.(bool); !ok {
					continue
				}
				if results[packageName] == nil {
					results[packageName] = make(map[string]interface{})
				}
				if r, ok := results[packageName].(map[string]interface{}); ok {
					r[k] = v
				}
			}
		}
	}

	paths, err := protoutil.NewStruct(results)
	if err != nil {
		return resp, err
	}

	resp = &authz.DecisionTreeResponse{
		PathRoot: req.PolicyContext.Path,
		Path:     paths,
	}

	return resp, nil
}

// Is decision eval function.
func (s *AuthorizerServer) Is(ctx context.Context, req *authz.IsRequest) (*authz.IsResponse, error) { // nolint:funlen,gocyclo //TODO: split into smaller functions after merge with onebox
	log := grpcutil.CompleteLogger(ctx, s.logger)

	resp := &authz.IsResponse{
		Decisions: make([]*authz.Decision, 0),
	}

	if req.PolicyContext == nil {
		return resp, cerr.ErrInvalidArgument.Msg("policy context not set")
	}

	if req.PolicyContext.GetId() == "" {
		return resp, cerr.ErrInvalidArgument.Msg("policy context id not set")
	}

	if req.PolicyContext.Path == "" {
		return resp, cerr.ErrInvalidArgument.Msg("policy context path not set")
	}

	if len(req.PolicyContext.Decisions) == 0 {
		return resp, cerr.ErrInvalidArgument.Msg("policy context decisions not set")
	}

	if req.ResourceContext == nil {
		req.ResourceContext = pb.NewStruct()
	}

	if req.IdentityContext == nil {
		return resp, cerr.ErrInvalidArgument.Msg("identity context not set")
	}

	if req.IdentityContext.Type == api.IdentityType_IDENTITY_TYPE_UNKNOWN {
		return resp, cerr.ErrInvalidArgument.Msg("identity type UNKNOWN")
	}

	user, err := s.getUserFromIdentityContext(ctx, req.IdentityContext)
	if err != nil {
		log.Error().Err(err).Interface("req", req).Msg("failed to resolve identity context")
		return resp, cerr.ErrUserNotFound.WithGRPCStatus(codes.NotFound).Msg("failed to resolve identity context")
	}

	input := map[string]interface{}{
		InputUser:     convert(user),
		InputIdentity: convert(req.IdentityContext),
		InputPolicy:   req.PolicyContext,
		InputResource: req.ResourceContext,
	}

	queryStmt := fmt.Sprintf("x = data.%s", req.PolicyContext.Path)

	log.Debug().Interface("input", input).Msg("calculating is")

	policyRuntime, err := s.runtimeResolver.RuntimeFromContext(ctx, req.PolicyContext.GetId(), req.PolicyContext.GetName(), req.PolicyContext.InstanceLabel)
	if err != nil {
		return resp, errors.Wrap(err, "failed to procure tenant runtime")
	}

	query, err := rego.New(
		rego.Compiler(policyRuntime.GetPluginsManager().GetCompiler()),
		rego.Store(policyRuntime.GetPluginsManager().Store),
		rego.Query(queryStmt),
	).PrepareForEval(ctx)

	if err != nil {
		return resp, cerr.ErrBadQuery.Err(err).Str("query", queryStmt)
	}

	results, err := query.Eval(ctx, rego.EvalInput(input))

	if err != nil {
		return resp, cerr.ErrBadQuery.Err(err).Str("query", queryStmt).Msg("query evaluation failed")
	} else if len(results) == 0 {
		return resp, cerr.ErrBadQuery.Err(err).Str("query", queryStmt).Msg("undefined results")
	}

	v := results[0].Bindings["x"]
	outcomes := map[string]bool{}

	for _, d := range req.PolicyContext.Decisions {
		decision := authz.Decision{
			Decision: d,
		}
		decision.Is, err = is(v, d)
		if err != nil {
			return nil, errors.Wrapf(err, "failed getting outcome for decision [%s]", d)
		}
		resp.Decisions = append(resp.Decisions, &decision)
		outcomes[decision.Decision] = decision.Is
	}

	dplugin := decisionlog_plugin.Lookup(policyRuntime.GetPluginsManager())
	d := dl.Decision{
		Id:        uuid.NewString(),
		Timestamp: timestamppb.New(time.Now().In(time.UTC)),
		TenantId:  ids.ExtractTenantID(ctx),
		Path:      req.PolicyContext.Path,
		Policy: &dl.DecisionPolicy{
			Context: req.PolicyContext,
		},
		User: &dl.DecisionUser{
			Context: req.IdentityContext,
			Id:      getID(input),
			Email:   getEmail(input),
		},
		Resource: req.ResourceContext,
		Outcomes: outcomes,
	}

	if dplugin == nil {
		return resp, err
	}

	err = dplugin.Log(ctx, &d)
	if err != nil {
		return resp, err
	}

	return resp, err
}

func is(v interface{}, decision string) (bool, error) {
	switch x := v.(type) {
	case bool:
		outcome := v.(bool)
		return outcome, nil
	case map[string]interface{}:
		m := v.(map[string]interface{})
		if _, ok := m[decision]; !ok {
			return false, cerr.ErrInvalidDecision.Msgf("decision element [%s] not found", decision)
		}
		outcome, err := is(m[decision], decision)
		if err != nil {
			return false, cerr.ErrInvalidDecision.Err(err)
		}
		return outcome, nil
	default:
		return false, cerr.ErrInvalidDecision.Msgf("is unexpected type %T", x)
	}
}

func (s *AuthorizerServer) Query(ctx context.Context, req *authz.QueryRequest) (*authz.QueryResponse, error) { // nolint:funlen,gocyclo //TODO: split into smaller functions after merge with onebox
	log := grpcutil.CompleteLogger(ctx, s.logger)

	if req.Query == "" {
		return &authz.QueryResponse{}, cerr.ErrInvalidArgument.Msg("query not set")
	}

	if req.Options == nil {
		req.Options = &authz.QueryOptions{
			Metrics:      false,
			Instrument:   false,
			Trace:        authz.TraceLevel_TRACE_LEVEL_OFF,
			TraceSummary: false,
		}
	}

	if req.Options.Trace == authz.TraceLevel_TRACE_LEVEL_UNKNOWN {
		req.Options.Trace = authz.TraceLevel_TRACE_LEVEL_OFF
	}

	var input map[string]interface{}

	if req.Input != "" {
		if err := json.Unmarshal([]byte(req.Input), &input); err != nil {
			return &authz.QueryResponse{}, errors.Wrap(err, "failed to unmarshal input - make sure it's a valid JSON object")
		}
	}

	if input == nil {
		input = make(map[string]interface{})
	}

	if req.PolicyContext != nil {
		input[InputPolicy] = req.PolicyContext
	}

	if req.ResourceContext != nil {
		input[InputResource] = req.ResourceContext
	}

	if req.IdentityContext != nil {
		if req.IdentityContext.Type == api.IdentityType_IDENTITY_TYPE_UNKNOWN {
			return &authz.QueryResponse{}, cerr.ErrInvalidArgument.Msg("identity type UNKNOWN")
		}

		if req.IdentityContext.Type != api.IdentityType_IDENTITY_TYPE_NONE {
			input[InputIdentity] = convert(req.IdentityContext)
		}
	}

	if req.IdentityContext != nil && req.IdentityContext.Type != api.IdentityType_IDENTITY_TYPE_NONE {
		user, err := s.getUserFromIdentityContext(ctx, req.IdentityContext)
		if err != nil || user == nil {
			if err != nil {
				log.Error().Err(err).Interface("req", req).Msg("failed to resolve identity context")
			}

			return &authz.QueryResponse{}, cerr.ErrAuthenticationFailed.WithGRPCStatus(codes.NotFound).Msg("failed to resolve identity context")
		}

		input[InputUser] = convert(user)
	}

	log.Debug().Str("query", req.Query).Interface("input", input).Msg("executing query")
	var rt *runtime.Runtime
	var err error
	if req.PolicyContext != nil {
		rt, err = s.runtimeResolver.RuntimeFromContext(ctx, req.PolicyContext.GetId(), req.PolicyContext.GetName(), req.PolicyContext.InstanceLabel)
		if err != nil {
			return &authz.QueryResponse{}, errors.Wrap(err, "failed to procure tenant runtime")
		}
	} else {
		return &authz.QueryResponse{}, cerr.ErrInvalidPolicyID.Msg("undefined policy context")
	}

	queryResult, err := rt.Query(
		ctx,
		req.Query,
		input,
		req.Options.TraceSummary,
		req.Options.Metrics,
		req.Options.Instrument,
		TraceLevelToExplainModeV1(req.Options.Trace),
	)
	if err != nil {
		return &authz.QueryResponse{}, err
	}

	resp := &authz.QueryResponse{}

	// results
	for _, result := range queryResult.Result {
		structValue, errX := protoutil.NewStruct(result.Bindings.WithoutWildcards())
		if errX != nil {
			return resp, errors.Wrap(err, "failed to create result grpc object")
		}

		resp.Results = append(resp.Results, structValue)
	}

	// metrics
	if queryResult.Metrics != nil {
		if metricsStruct, errX := protoutil.NewStruct(queryResult.Metrics); errX == nil {
			resp.Metrics = metricsStruct
		}
	} else {
		resp.Metrics, _ = protoutil.NewStruct(make(map[string]interface{}))
	}

	// trace (explanation)
	if queryResult.Explanation != nil {
		var v []interface{}
		if err = json.Unmarshal(queryResult.Explanation, &v); err != nil {
			return resp, errors.Wrap(err, "unmarshal json")
		}

		list, err := protoutil.NewList(v)
		if err != nil {
			rt.Logger.Error().Err(err).Msg("newList")
		}

		if req.Options.TraceSummary {
			for _, val := range list.Values {
				resp.TraceSummary = append(resp.TraceSummary, val.GetStringValue())
			}
		} else {
			for _, val := range list.Values {
				resp.Trace = append(resp.Trace, val.GetStructValue())
			}
		}
	}

	return resp, nil
}

// convert, explicitly convert from proto message interface{} in order
// to preserve enum values as strings when marshaled to JSON
func convert(msg proto.Message) interface{} {
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

	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return nil
	}

	return v
}

func TraceLevelToExplainModeV1(t authz.TraceLevel) types.ExplainModeV1 {
	switch t {
	case authz.TraceLevel_TRACE_LEVEL_UNKNOWN:
		return types.ExplainOffV1
	case authz.TraceLevel_TRACE_LEVEL_OFF:
		return types.ExplainOffV1
	case authz.TraceLevel_TRACE_LEVEL_FULL:
		return types.ExplainFullV1
	case authz.TraceLevel_TRACE_LEVEL_NOTES:
		return types.ExplainNotesV1
	case authz.TraceLevel_TRACE_LEVEL_FAILS:
		return types.ExplainFailsV1
	default:
		return types.ExplainOffV1
	}
}

func initDecisionFilter(decisions []string) func(decision string) bool {
	if len(decisions) == 1 && decisions[0] == "*" {
		return func(s string) bool {
			return true
		}
	}

	decisionMap := make(map[string]struct{})
	for _, v := range decisions {
		decisionMap[v] = struct{}{}
	}

	return func(s string) bool {
		_, ok := decisionMap[s]
		return ok
	}
}

func getID(v map[string]interface{}) string {
	if u, ok := v["user"].(map[string]interface{}); ok {
		if i, ok := u["id"].(string); ok {
			return i
		}
	}
	return ""
}

func getEmail(v map[string]interface{}) string {
	if u, ok := v["user"].(map[string]interface{}); ok {
		if e, ok := u["email"].(string); ok {
			return e
		}
		if p, ok := u["properties"].(map[string]interface{}); ok {
			if e, ok := p["email"].(string); ok {
				return e
			}
		}
	}
	return ""
}
