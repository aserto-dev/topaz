package impl

import (
	"context"
	"encoding/json"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	"github.com/aserto-dev/go-directory/pkg/pb"
	runtime "github.com/aserto-dev/runtime"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *AuthorizerServer) Compile(ctx context.Context, req *authorizer.CompileRequest) (*authorizer.CompileResponse, error) {
	log := s.logger.With().Str("api", "compile").Logger()

	if err := s.compileVerifyRequest(req); err != nil {
		return &authorizer.CompileResponse{}, err
	}

	input, err := s.compileSetInput(ctx, req)
	if err != nil {
		return &authorizer.CompileResponse{}, err
	}

	log.Debug().Str("compile", req.GetQuery()).Interface("input", input).Msg("compile")

	rt, err := s.getRuntime(ctx, req.GetPolicyInstance())
	if err != nil {
		return &authorizer.CompileResponse{}, err
	}

	if _, err = rt.ValidateQuery(req.GetQuery()); err != nil {
		return &authorizer.CompileResponse{}, aerr.ErrBadQuery.Err(err)
	}

	compileResult, err := rt.Compile(ctx, req.GetQuery(),
		input,
		req.GetUnknowns(),
		req.GetDisableInlining(),
		true,
		req.GetOptions().GetMetrics(),
		req.GetOptions().GetInstrument(),
		traceLevelToExplainModeV2(req.GetOptions().GetTrace()))
	if err != nil {
		return &authorizer.CompileResponse{}, err
	}

	resp := &authorizer.CompileResponse{}

	resp.Result, err = s.compileSetResult(compileResult)
	if err != nil {
		return &authorizer.CompileResponse{}, err
	}

	resp.Metrics, err = s.compileSetMetrics(compileResult)
	if err != nil {
		return &authorizer.CompileResponse{}, err
	}

	if req.GetOptions().GetTrace() > authorizer.TraceLevel_TRACE_LEVEL_OFF {
		resp.Trace, err = s.compileSetTrace(compileResult)
		if err != nil {
			return &authorizer.CompileResponse{}, err
		}
	}

	if req.GetOptions().GetTraceSummary() {
		resp.TraceSummary, err = s.compileSetTraceSummary(compileResult)
		if err != nil {
			return &authorizer.CompileResponse{}, err
		}
	}

	return resp, nil
}

func (*AuthorizerServer) compileVerifyRequest(req *authorizer.CompileRequest) error {
	if req.GetQuery() == "" {
		return aerr.ErrInvalidArgument.Msg("query not set")
	}

	if req.GetIdentityContext() == nil || req.GetIdentityContext().GetType() == api.IdentityType_IDENTITY_TYPE_UNKNOWN {
		return aerr.ErrInvalidArgument.Msg("identity type UNKNOWN")
	}

	if req.GetOptions() == nil {
		req.Options = &authorizer.QueryOptions{
			Metrics:      false,
			Instrument:   false,
			Trace:        authorizer.TraceLevel_TRACE_LEVEL_OFF,
			TraceSummary: false,
		}
	}

	if req.GetOptions().GetTrace() == authorizer.TraceLevel_TRACE_LEVEL_UNKNOWN {
		req.Options.Trace = authorizer.TraceLevel_TRACE_LEVEL_OFF
	}

	return nil
}

func (s *AuthorizerServer) compileSetInput(ctx context.Context, req *authorizer.CompileRequest) (map[string]any, error) {
	input := map[string]any{}

	if req.GetInput() != "" {
		if err := json.Unmarshal([]byte(req.GetInput()), &input); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal input - make sure it's a valid JSON object")
		}
	}

	if req.GetResourceContext() != nil {
		input[InputResource] = req.GetResourceContext()
	}

	if err := s.resolveIdentityContext(ctx, req.GetIdentityContext(), input); err != nil {
		return nil, err
	}

	if req.GetPolicyContext() != nil {
		input[InputPolicy] = req.GetPolicyContext()
	}

	return input, nil
}

func (s *AuthorizerServer) compileSetResult(compileResult *runtime.CompileResult) (*structpb.Struct, error) {
	compileResultJSON, err := json.Marshal(compileResult.Result)
	if err != nil {
		return nil, err
	}

	var compileResultMap map[string]any
	if err := json.Unmarshal(compileResultJSON, &compileResultMap); err != nil {
		return nil, err
	}

	result, err := structpb.NewStruct(compileResultMap)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *AuthorizerServer) compileSetMetrics(compileResult *runtime.CompileResult) (*structpb.Struct, error) {
	if compileResult.Metrics == nil {
		return pb.NewStruct(), nil
	}

	metricsStruct, err := structpb.NewStruct(compileResult.Metrics)
	if err != nil {
		return nil, err
	}

	return metricsStruct, nil
}

func (s *AuthorizerServer) compileSetTrace(compileResult *runtime.CompileResult) ([]*structpb.Struct, error) {
	list, err := s.traceList(compileResult)
	if err != nil {
		return []*structpb.Struct{}, err
	}

	result := []*structpb.Struct{}
	for _, val := range list.GetValues() {
		result = append(result, val.GetStructValue())
	}

	return result, nil
}

func (s *AuthorizerServer) compileSetTraceSummary(compileResult *runtime.CompileResult) ([]string, error) {
	list, err := s.traceList(compileResult)
	if err != nil {
		return []string{}, err
	}

	result := []string{}
	for _, val := range list.GetValues() {
		result = append(result, val.GetStringValue())
	}

	return result, nil
}

func (*AuthorizerServer) traceList(compileResult *runtime.CompileResult) (*structpb.ListValue, error) {
	if compileResult.Explanation == nil {
		return &structpb.ListValue{}, nil
	}

	var v []any
	if err := json.Unmarshal(compileResult.Explanation, &v); err != nil {
		return nil, errors.Wrap(err, "unmarshal json")
	}

	list, err := structpb.NewList(v)
	if err != nil {
		return nil, errors.Wrap(err, "creating list")
	}

	return list, nil
}
