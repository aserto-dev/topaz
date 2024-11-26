package impl

import (
	"context"
	"encoding/json"
	"fmt"
	goruntime "runtime"
	"strings"
	"sync"
	"time"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	"github.com/aserto-dev/go-directory/pkg/pb"
	"github.com/aserto-dev/header"
	runtime "github.com/aserto-dev/runtime"
	decisionlog_plugin "github.com/aserto-dev/topaz/plugins/decision_log"

	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/version"
	"github.com/aserto-dev/topaz/resolvers"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/mennanov/fmutils"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/server/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	jwkCache *jwk.Cache

	resolver *resolvers.Resolvers
}

func NewAuthorizerServer(
	ctx context.Context,
	logger *zerolog.Logger,
	cfg *config.Common,
	rf *resolvers.Resolvers,
) *AuthorizerServer {
	newLogger := logger.With().Str("component", "api.grpc").Logger()

	jwkCache := jwk.NewCache(ctx)

	return &AuthorizerServer{
		cfg:      cfg,
		logger:   &newLogger,
		resolver: rf,
		jwkCache: jwkCache,
	}
}

func (s *AuthorizerServer) DecisionTree(ctx context.Context, req *authorizer.DecisionTreeRequest) (*authorizer.DecisionTreeResponse, error) { // nolint:funlen,gocyclo //TODO: split into smaller functions after merge with onebox
	log := s.logger.With().Str("api", "decision_tree").Logger()

	resp := &authorizer.DecisionTreeResponse{}

	if req.PolicyContext == nil {
		return resp, aerr.ErrInvalidArgument.Msg("policy context not set")
	}

	if req.ResourceContext == nil {
		req.ResourceContext = &structpb.Struct{Fields: make(map[string]*structpb.Value)}
	}

	if req.Options == nil {
		req.Options = &authorizer.DecisionTreeOptions{
			PathSeparator: authorizer.PathSeparator_PATH_SEPARATOR_DOT,
		}
	}

	if req.IdentityContext == nil {
		return resp, aerr.ErrInvalidArgument.Msg("identity context not set")
	}

	if req.IdentityContext.Type == api.IdentityType_IDENTITY_TYPE_UNKNOWN {
		return resp, aerr.ErrInvalidArgument.Msg("identity type UNKNOWN")
	}

	user, err := s.getUserFromIdentityContext(ctx, req.IdentityContext)
	if err != nil {
		log.Error().Err(err).Interface("req", req).Msg("failed to resolve identity context")
		return resp, aerr.ErrAuthenticationFailed.WithGRPCStatus(codes.NotFound).Msg("failed to resolve identity context")
	}

	input := map[string]interface{}{
		InputUser:     convert(user),
		InputIdentity: convert(req.IdentityContext),
		InputPolicy:   req.PolicyContext,
		InputResource: req.ResourceContext,
	}

	policyRuntime, err := s.getRuntime(ctx, req.PolicyInstance)
	if err != nil {
		return resp, err
	}

	listPolicies, err := policyRuntime.ListPolicies(ctx)
	if err != nil {
		return resp, errors.Wrap(err, "get policy list")
	}

	decisionFilter := initDecisionFilter(req.PolicyContext.Decisions)

	queryStmt := strings.Builder{}
	r := 0
	for i := 0; i < len(listPolicies); i++ {
		for _, rule := range listPolicies[i].AST.Rules {
			if !strings.HasPrefix(listPolicies[i].AST.Package.Path.String(), "data."+req.PolicyContext.Path) {
				continue
			}

			if !decisionFilter(rule.Head.Name.String()) {
				continue
			}

			queryStmt.WriteString(fmt.Sprintf("r%d = %s\n", i, listPolicies[i].AST.Package.Path))
			r++
			break
		}
	}

	if queryStmt.Len() == 0 {
		return resp, aerr.ErrInvalidArgument.Msg("no decisions specified")
	}

	results := make(map[string]interface{})

	qry, err := rego.New(
		rego.Compiler(policyRuntime.GetPluginsManager().GetCompiler()),
		rego.Store(policyRuntime.GetPluginsManager().Store),
		rego.Input(input),
		rego.Query(queryStmt.String()),
	).PrepareForEval(ctx)
	if err != nil {
		return resp, aerr.ErrBadQuery.Err(err).Msg(queryStmt.String())
	}

	queryResults, err := qry.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return resp, aerr.ErrBadQuery.Err(err).Msgf("query evaluation failed: %s", queryStmt.String())
	} else if len(queryResults) == 0 {
		return resp, aerr.ErrBadQuery.Err(err).Msgf("undefined results: %s", queryStmt.String())
	}

	for _, expression := range queryResults[0].Expressions {
		expr := strings.Split(expression.Text, "=")
		rule := strings.TrimSpace(expr[0])
		path := strings.TrimPrefix(strings.TrimSpace(expr[1]), "data.")
		if req.Options.PathSeparator == authorizer.PathSeparator_PATH_SEPARATOR_SLASH {
			path = strings.ReplaceAll(path, ".", "/")
		}

		binding, ok := queryResults[0].Bindings[rule].(map[string]interface{})
		if !ok {
			continue
		}

		outcomes := make(map[string]interface{})
		for _, decision := range req.PolicyContext.Decisions {
			if r, ok := binding[decision].(bool); ok {
				outcomes[decision] = r
			}
		}

		results[path] = outcomes
	}

	paths, err := structpb.NewStruct(results)
	if err != nil {
		return resp, err
	}

	resp = &authorizer.DecisionTreeResponse{
		PathRoot: req.PolicyContext.Path,
		Path:     paths,
	}

	return resp, nil
}

// Is decision eval function.
func (s *AuthorizerServer) Is(ctx context.Context, req *authorizer.IsRequest) (*authorizer.IsResponse, error) { // nolint:funlen,gocyclo //TODO: split into smaller functions after merge with onebox
	log := s.logger.With().Str("api", "is").Logger()

	resp := &authorizer.IsResponse{
		Decisions: []*authorizer.Decision{},
	}

	if req.PolicyContext == nil {
		return resp, aerr.ErrInvalidArgument.Msg("policy context not set")
	}

	if req.PolicyContext.Path == "" {
		return resp, aerr.ErrInvalidArgument.Msg("policy context path not set")
	}

	if len(req.PolicyContext.Decisions) == 0 {
		return resp, aerr.ErrInvalidArgument.Msg("policy context decisions not set")
	}

	if req.ResourceContext == nil {
		req.ResourceContext = pb.NewStruct()
	}

	if req.IdentityContext == nil {
		return resp, aerr.ErrInvalidArgument.Msg("identity context not set")
	}

	if req.IdentityContext.Type == api.IdentityType_IDENTITY_TYPE_UNKNOWN {
		return resp, aerr.ErrInvalidArgument.Msg("identity type UNKNOWN")
	}

	user, err := s.getUserFromIdentityContext(ctx, req.IdentityContext)
	if err != nil {
		log.Error().Err(err).Interface("req", req).Msg("failed to resolve identity context")
		return resp, aerr.ErrUserNotFound.WithGRPCStatus(codes.NotFound).Msg("failed to resolve identity context")
	}

	input := map[string]interface{}{
		InputUser:     convert(user),
		InputIdentity: convert(req.IdentityContext),
		InputPolicy:   req.PolicyContext,
		InputResource: req.ResourceContext,
	}

	log.Debug().Interface("input", input).Msg("is")

	policyRuntime, err := s.getRuntime(ctx, req.PolicyInstance)
	if err != nil {
		return resp, err
	}

	sb := strings.Builder{}
	for i, decision := range req.PolicyContext.Decisions {
		sb.WriteString(fmt.Sprintf("x%d = data.%s.%s\n", i, req.PolicyContext.Path, decision))
	}

	queryStmt := sb.String()

	query, err := rego.New(
		rego.Compiler(policyRuntime.GetPluginsManager().GetCompiler()),
		rego.Store(policyRuntime.GetPluginsManager().Store),
		rego.Query(queryStmt),
	).PrepareForEval(ctx)
	if err != nil {
		return resp, aerr.ErrBadQuery.Err(err).Msg(queryStmt)
	}

	results, err := query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return resp, aerr.ErrBadQuery.Err(err).Msgf("query evaluation failed: %s", queryStmt)
	}
	if len(results) == 0 {
		return resp, aerr.ErrBadQuery.Err(err).Msgf("undefined results: %s", queryStmt)
	}

	for i, d := range req.PolicyContext.Decisions {
		v, ok := results[0].Bindings[fmt.Sprintf("x%d", i)]
		if !ok {
			return nil, errors.Wrapf(err, "failed getting binding for decision [%s]", d)
		}

		outcome, ok := v.(bool)
		if !ok {
			return nil, errors.Wrapf(err, "non-boolean outcome for decision [%s]: %s", d, v)
		}

		decision := authorizer.Decision{
			Decision: d,
			Is:       outcome,
		}

		resp.Decisions = append(resp.Decisions, &decision)
	}

	dlPlugin := decisionlog_plugin.Lookup(policyRuntime.GetPluginsManager())
	if dlPlugin == nil {
		return resp, err
	}

	d := api.Decision{
		Id:        uuid.NewString(),
		Timestamp: timestamppb.New(time.Now().In(time.UTC)),
		Path:      req.PolicyContext.Path,
		Policy: &api.DecisionPolicy{
			Context:        req.PolicyContext,
			PolicyInstance: req.PolicyInstance,
		},
		User: &api.DecisionUser{
			Context: req.IdentityContext,
			Id:      getID(input),
			Email:   getEmail(input),
		},
		TenantId: getTenantID(ctx),
		Resource: req.ResourceContext,
		Outcomes: getOutcomes(resp.Decisions),
	}

	err = dlPlugin.Log(ctx, &d)
	if err != nil {
		return resp, err
	}

	return resp, err
}

func getTenantID(ctx context.Context) *string {
	tenantID := header.ExtractTenantID(ctx)
	if tenantID != "" {
		return &tenantID
	}
	return nil
}

func getOutcomes(decisions []*authorizer.Decision) map[string]bool {
	return lo.SliceToMap(decisions, func(item *authorizer.Decision) (string, bool) {
		return item.Decision, item.Is
	})
}

func (s *AuthorizerServer) Query(ctx context.Context, req *authorizer.QueryRequest) (*authorizer.QueryResponse, error) { // nolint:funlen,gocyclo //TODO: split into smaller functions after merge with onebox
	log := s.logger.With().Str("api", "query").Logger()

	if req.Query == "" {
		return &authorizer.QueryResponse{}, aerr.ErrInvalidArgument.Msg("query not set")
	}

	if req.Options == nil {
		req.Options = &authorizer.QueryOptions{
			Metrics:      false,
			Instrument:   false,
			Trace:        authorizer.TraceLevel_TRACE_LEVEL_OFF,
			TraceSummary: false,
		}
	}

	if req.Options.Trace == authorizer.TraceLevel_TRACE_LEVEL_UNKNOWN {
		req.Options.Trace = authorizer.TraceLevel_TRACE_LEVEL_OFF
	}

	var input map[string]interface{}

	if req.Input != "" {
		if err := json.Unmarshal([]byte(req.Input), &input); err != nil {
			return &authorizer.QueryResponse{}, errors.Wrap(err, "failed to unmarshal input - make sure it's a valid JSON object")
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
			return &authorizer.QueryResponse{}, aerr.ErrInvalidArgument.Msg("identity type UNKNOWN")
		}

		if req.IdentityContext.Type != api.IdentityType_IDENTITY_TYPE_NONE {
			input[InputIdentity] = convert(req.IdentityContext)
		}

		if req.IdentityContext.Type != api.IdentityType_IDENTITY_TYPE_NONE {
			user, err := s.getUserFromIdentityContext(ctx, req.IdentityContext)
			if err != nil || user == nil {
				if err != nil {
					log.Error().Err(err).Interface("req", req).Msg("failed to resolve identity context")
				}

				return &authorizer.QueryResponse{}, aerr.ErrAuthenticationFailed.WithGRPCStatus(codes.NotFound).Msg("failed to resolve identity context")
			}

			input[InputUser] = convert(user)
		}
	}

	log.Debug().Str("query", req.Query).Interface("input", input).Msg("query")

	rt, err := s.getRuntime(ctx, req.PolicyInstance)
	if err != nil {
		return &authorizer.QueryResponse{}, err
	}

	_, err = rt.ValidateQuery(req.Query)
	if err != nil {
		return &authorizer.QueryResponse{}, aerr.ErrBadQuery.Err(err)
	}

	queryResult, err := rt.Query(
		ctx,
		req.Query,
		input,
		req.Options.TraceSummary,
		req.Options.Metrics,
		req.Options.Instrument,
		TraceLevelToExplainModeV2(req.Options.Trace),
	)
	if err != nil {
		return &authorizer.QueryResponse{}, err
	}

	resp := &authorizer.QueryResponse{}
	queryResultJSON, err := json.Marshal(queryResult.Result)
	if err != nil {
		return resp, err
	}

	var queryResultMap []interface{}
	err = json.Unmarshal(queryResultJSON, &queryResultMap)
	if err != nil {
		return resp, err
	}
	respMap := make(map[string]interface{})
	respMap["result"] = queryResultMap
	resp.Response, err = structpb.NewStruct(respMap)
	if err != nil {
		return resp, err
	}

	// metrics
	if queryResult.Metrics != nil {
		if metricsStruct, errX := structpb.NewStruct(queryResult.Metrics); errX == nil {
			resp.Metrics = metricsStruct
		}
	} else {
		resp.Metrics, _ = structpb.NewStruct(make(map[string]interface{}))
	}

	// trace (explanation)
	if queryResult.Explanation != nil {
		var v []interface{}
		if err = json.Unmarshal(queryResult.Explanation, &v); err != nil {
			return resp, errors.Wrap(err, "unmarshal json")
		}

		list, err := structpb.NewList(v)
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

// nolint: staticcheck
func (s *AuthorizerServer) getRuntime(ctx context.Context, policyInstance *api.PolicyInstance) (*runtime.Runtime, error) {
	var rt *runtime.Runtime
	var err error
	if policyInstance != nil {
		rt, err = s.resolver.GetRuntimeResolver().RuntimeFromContext(ctx, policyInstance.Name)
		if err != nil {
			return nil, errors.Wrap(err, "failed to procure tenant runtime")
		}
	} else {
		rt, err = s.resolver.GetRuntimeResolver().RuntimeFromContext(ctx, "")
		if err != nil {
			return nil, aerr.ErrInvalidPolicyID.Msg("undefined policy context")
		}
	}
	return rt, err
}

func (s *AuthorizerServer) Compile(ctx context.Context, req *authorizer.CompileRequest) (*authorizer.CompileResponse, error) { // nolint:funlen,gocyclo //TODO: split into smaller functions after merge with onebox
	log := s.logger.With().Str("api", "compile").Logger()

	if req.Query == "" {
		return &authorizer.CompileResponse{}, aerr.ErrInvalidArgument.Msg("query not set")
	}

	if req.Options == nil {
		req.Options = &authorizer.QueryOptions{
			Metrics:      false,
			Instrument:   false,
			Trace:        authorizer.TraceLevel_TRACE_LEVEL_OFF,
			TraceSummary: false,
		}
	}

	if req.Options.Trace == authorizer.TraceLevel_TRACE_LEVEL_UNKNOWN {
		req.Options.Trace = authorizer.TraceLevel_TRACE_LEVEL_OFF
	}

	var input map[string]interface{}

	if req.Input != "" {
		if err := json.Unmarshal([]byte(req.Input), &input); err != nil {
			return &authorizer.CompileResponse{}, errors.Wrap(err, "failed to unmarshal input - make sure it's a valid JSON object")
		}
	}

	if input == nil {
		input = make(map[string]interface{})
	}

	if req.ResourceContext != nil {
		input[InputResource] = req.ResourceContext
	}

	if req.IdentityContext != nil {
		if req.IdentityContext.Type == api.IdentityType_IDENTITY_TYPE_UNKNOWN {
			return &authorizer.CompileResponse{}, aerr.ErrInvalidArgument.Msg("identity type UNKNOWN")
		}

		if req.IdentityContext.Type != api.IdentityType_IDENTITY_TYPE_NONE {
			input[InputIdentity] = convert(req.IdentityContext)
		}

		if req.IdentityContext.Type != api.IdentityType_IDENTITY_TYPE_NONE {
			user, err := s.getUserFromIdentityContext(ctx, req.IdentityContext)
			if err != nil || user == nil {
				if err != nil {
					log.Error().Err(err).Interface("req", req).Msg("failed to resolve identity context")
				}

				return &authorizer.CompileResponse{}, aerr.ErrAuthenticationFailed.WithGRPCStatus(codes.NotFound).Msg("failed to resolve identity context")
			}

			input[InputUser] = convert(user)
		}
	}

	if req.PolicyContext != nil {
		input[InputPolicy] = req.PolicyContext
	}

	if input == nil {
		input = make(map[string]interface{})
	}
	log.Debug().Str("compile", req.Query).Interface("input", input).Msg("compile")
	rt, err := s.getRuntime(ctx, req.PolicyInstance)
	if err != nil {
		return &authorizer.CompileResponse{}, err
	}

	_, err = rt.ValidateQuery(req.Query)
	if err != nil {
		return &authorizer.CompileResponse{}, aerr.ErrBadQuery.Err(err)
	}

	compileResult, err := rt.Compile(ctx, req.Query,
		input,
		req.Unknowns,
		req.DisableInlining,
		true,
		req.Options.Metrics,
		req.Options.Instrument,
		TraceLevelToExplainModeV2(req.Options.Trace))
	resp := &authorizer.CompileResponse{}
	if err != nil {
		return resp, err
	}
	compileResultJSON, err := json.Marshal(compileResult.Result)
	if err != nil {
		return resp, err
	}

	var compileResultMap map[string]interface{}
	err = json.Unmarshal(compileResultJSON, &compileResultMap)
	if err != nil {
		return resp, err
	}
	resp.Result, err = structpb.NewStruct(compileResultMap)
	if err != nil {
		return resp, err
	}
	// metrics
	if compileResult.Metrics != nil {
		if metricsStruct, errX := structpb.NewStruct(compileResult.Metrics); errX == nil {
			resp.Metrics = metricsStruct
		}
	} else {
		resp.Metrics, _ = structpb.NewStruct(make(map[string]interface{}))
	}

	// trace (explanation)
	if compileResult.Explanation != nil {
		var v []interface{}
		if err = json.Unmarshal(compileResult.Explanation, &v); err != nil {
			return resp, errors.Wrap(err, "unmarshal json")
		}

		list, err := structpb.NewList(v)
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

func (s *AuthorizerServer) ListPolicies(ctx context.Context, req *authorizer.ListPoliciesRequest) (*authorizer.ListPoliciesResponse, error) {
	response := &authorizer.ListPoliciesResponse{}

	rt, err := s.getRuntime(ctx, req.PolicyInstance)
	if err != nil {
		return response, errors.Wrap(err, "failed to get runtime")
	}

	policies, err := rt.ListPolicies(ctx)
	if err != nil {
		return response, err
	}

	for _, policy := range policies {

		module, err := policyToModule(policy)
		if err != nil {
			return response, errors.Wrapf(err, "failed to parse policy with ID [%s]", policy.ID)
		}

		if req.FieldMask != nil {
			paths := s.validateMask(req.FieldMask, &api.Module{})
			mask := fmutils.NestedMaskFromPaths(paths)
			mask.Filter(module)
		}

		response.Result = append(response.Result, module)
	}

	return response, nil
}

func (s *AuthorizerServer) GetPolicy(ctx context.Context, req *authorizer.GetPolicyRequest) (*authorizer.GetPolicyResponse, error) {
	response := &authorizer.GetPolicyResponse{}
	rt, err := s.getRuntime(ctx, req.PolicyInstance)
	if err != nil {
		return response, errors.Wrap(err, "failed to get runtime")
	}

	policy, err := rt.GetPolicy(ctx, req.Id)
	if err != nil {
		return response, errors.Wrapf(err, "failed to get policy with ID [%s]", req.Id)
	}

	if policy == nil {
		// TODO: add cerr
		return response, errors.Errorf("policy with ID [%s] not found", req.Id)
	}

	module, err := policyToModule(*policy)
	if err != nil {
		return response, errors.Wrap(err, "failed to convert policy to api.module")
	}

	if req.FieldMask != nil {
		paths := s.validateMask(req.FieldMask, &api.Module{})
		mask := fmutils.NestedMaskFromPaths(paths)
		mask.Filter(module)
	}

	response.Result = module
	return response, nil
}

func policyToModule(policy types.PolicyV1) (*api.Module, error) {
	astBts, err := json.Marshal(policy.AST)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal AST")
	}

	var v interface{}
	err = json.Unmarshal(astBts, &v)
	if err != nil {
		return nil, errors.Wrap(err, "failed to determine AST")
	}

	astValue, err := structpb.NewValue(v)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create astValue")
	}
	var packageName string
	if policy.AST != nil {
		packageName = policy.AST.Package.Path.String()
	}
	module := api.Module{
		Id:          &policy.ID,
		Raw:         &policy.Raw,
		PackagePath: &packageName,
		Ast:         astValue,
	}
	return &module, nil
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

func TraceLevelToExplainModeV2(t authorizer.TraceLevel) types.ExplainModeV1 {
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

// convert, explicitly converts from proto message interface{} in order
// to preserve enum values as strings when marshaled to JSON.
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

// validateMask checks if provided mask is validate.
func (s *AuthorizerServer) validateMask(mask *fieldmaskpb.FieldMask, protomsg protoreflect.ProtoMessage) []string {
	if len(mask.Paths) > 0 && mask.Paths[0] == "" {
		return []string{}
	}

	mask.Normalize()

	if !mask.IsValid(protomsg) {
		s.logger.Error().Msgf("field mask invalid %q", mask.Paths)
		return []string{}
	}

	return mask.Paths
}
