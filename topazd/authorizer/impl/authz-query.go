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

func (s *AuthorizerServer) Query(ctx context.Context, req *authorizer.QueryRequest) (*authorizer.QueryResponse, error) {
	log := s.logger.With().Str("api", "query").Logger()

	if err := s.queryVerifyRequest(req); err != nil {
		return &authorizer.QueryResponse{}, err
	}

	input, err := s.querySetInput(ctx, req)
	if err != nil {
		return &authorizer.QueryResponse{}, err
	}

	log.Debug().Str("query", req.GetQuery()).Interface("input", input).Msg("query")

	rt, err := s.getRuntime(ctx, s.instanceName(req))
	if err != nil {
		return &authorizer.QueryResponse{}, err
	}

	if _, err := rt.ValidateQuery(req.GetQuery()); err != nil {
		return &authorizer.QueryResponse{}, aerr.ErrBadQuery.Err(err)
	}

	queryResult, err := rt.Query(
		ctx,
		req.GetQuery(),
		input,
		req.GetOptions().GetTraceSummary(),
		req.GetOptions().GetMetrics(),
		req.GetOptions().GetInstrument(),
		traceLevelToExplainModeV2(req.GetOptions().GetTrace()),
	)
	if err != nil {
		return &authorizer.QueryResponse{}, err
	}

	resp := &authorizer.QueryResponse{}

	resp.Response, err = s.querySetResult(queryResult)
	if err != nil {
		return &authorizer.QueryResponse{}, err
	}

	resp.Metrics, err = s.querySetMetrics(queryResult)
	if err != nil {
		return &authorizer.QueryResponse{}, err
	}

	if req.GetOptions().GetTrace() > authorizer.TraceLevel_TRACE_LEVEL_OFF {
		resp.Trace, err = s.querySetTrace(queryResult)
		if err != nil {
			return &authorizer.QueryResponse{}, err
		}
	}

	if req.GetOptions().GetTraceSummary() {
		resp.TraceSummary, err = s.querySetTraceSummary(queryResult)
		if err != nil {
			return &authorizer.QueryResponse{}, err
		}
	}

	return resp, nil
}

func (*AuthorizerServer) queryVerifyRequest(req *authorizer.QueryRequest) error {
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

func (s *AuthorizerServer) querySetInput(ctx context.Context, req *authorizer.QueryRequest) (map[string]any, error) {
	input := map[string]any{}

	if req.GetInput() != "" {
		if err := json.Unmarshal([]byte(req.GetInput()), &input); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal input - make sure it's a valid JSON object")
		}
	}

	if err := s.resolveIdentityContext(ctx, req.GetIdentityContext(), input); err != nil {
		return nil, err
	}

	if req.GetPolicyContext() != nil {
		input[InputPolicy] = req.GetPolicyContext()
	}

	if req.GetResourceContext() != nil {
		input[InputResource] = req.GetResourceContext()
	}

	return input, nil
}

func (s *AuthorizerServer) querySetResult(queryResult *runtime.Result) (*structpb.Struct, error) {
	queryResultJSON, err := json.Marshal(queryResult.Result)
	if err != nil {
		return nil, err
	}

	var queryResultMap []any
	if err := json.Unmarshal(queryResultJSON, &queryResultMap); err != nil {
		return nil, err
	}

	respMap := map[string]any{
		"result": queryResultMap,
	}

	result, err := structpb.NewStruct(respMap)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *AuthorizerServer) querySetMetrics(queryResult *runtime.Result) (*structpb.Struct, error) {
	if queryResult.Metrics == nil {
		return pb.NewStruct(), nil
	}

	metricsStruct, err := structpb.NewStruct(queryResult.Metrics)
	if err != nil {
		return nil, err
	}

	return metricsStruct, nil
}

func (s *AuthorizerServer) querySetTrace(queryResult *runtime.Result) ([]*structpb.Struct, error) {
	list, err := s.queryList(queryResult)
	if err != nil {
		return []*structpb.Struct{}, err
	}

	result := []*structpb.Struct{}
	for _, val := range list.GetValues() {
		result = append(result, val.GetStructValue())
	}

	return result, nil
}

func (s *AuthorizerServer) querySetTraceSummary(queryResult *runtime.Result) ([]string, error) {
	list, err := s.queryList(queryResult)
	if err != nil {
		return []string{}, err
	}

	result := []string{}
	for _, val := range list.GetValues() {
		result = append(result, val.GetStringValue())
	}

	return result, nil
}

func (*AuthorizerServer) queryList(queryResult *runtime.Result) (*structpb.ListValue, error) {
	if queryResult.Explanation == nil {
		return &structpb.ListValue{}, nil
	}

	var v []any
	if err := json.Unmarshal(queryResult.Explanation, &v); err != nil {
		return nil, errors.Wrap(err, "unmarshal json")
	}

	list, err := structpb.NewList(v)
	if err != nil {
		return nil, errors.Wrap(err, "creating list")
	}

	return list, nil
}
