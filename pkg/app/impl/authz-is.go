package impl

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	"github.com/aserto-dev/go-directory/pkg/pb"
	"github.com/aserto-dev/header"
	decisionlog_plugin "github.com/aserto-dev/topaz/plugins/decisionlog"

	"github.com/google/uuid"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//nolint:funlen
func (s *AuthorizerServer) Is(ctx context.Context, req *authorizer.IsRequest) (*authorizer.IsResponse, error) {
	log := s.logger.With().Str("api", "is").Logger()

	if err := s.isVerifyRequest(req); err != nil {
		return &authorizer.IsResponse{}, err
	}

	input, err := s.isSetInput(ctx, req)
	if err != nil {
		return &authorizer.IsResponse{}, err
	}

	log.Debug().Interface("input", input).Msg("is")

	policyRuntime, err := s.getRuntime(ctx, req.GetPolicyInstance())
	if err != nil {
		return &authorizer.IsResponse{}, err
	}

	queryStmt := strings.Builder{}
	for i, decision := range req.GetPolicyContext().GetDecisions() {
		queryStmt.WriteString(fmt.Sprintf("x%d = data.%s.%s\n", i, req.GetPolicyContext().GetPath(), decision))
	}

	query, err := rego.New(
		rego.Compiler(policyRuntime.GetPluginsManager().GetCompiler()),
		rego.Store(policyRuntime.GetPluginsManager().Store),
		rego.Query(queryStmt.String()),
	).PrepareForEval(ctx)
	if err != nil {
		return &authorizer.IsResponse{}, aerr.ErrBadQuery.Err(err).Msg(queryStmt.String())
	}

	resp := &authorizer.IsResponse{
		Decisions: []*authorizer.Decision{},
	}

	queryResults, err := query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return resp, aerr.ErrBadQuery.Err(err).Msgf("query evaluation failed: %s", queryStmt.String())
	}

	if len(queryResults) == 0 {
		return resp, aerr.ErrBadQuery.Err(err).Msgf("undefined results: %s", queryStmt.String())
	}

	for i, d := range req.GetPolicyContext().GetDecisions() {
		v, ok := queryResults[0].Bindings[fmt.Sprintf("x%d", i)]
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

		resp.Decisions = append(resp.GetDecisions(), &decision)
	}

	dlPlugin := decisionlog_plugin.Lookup(policyRuntime.GetPluginsManager())
	if dlPlugin == nil {
		return resp, err
	}

	d := api.Decision{
		Id:        uuid.NewString(),
		Timestamp: timestamppb.New(time.Now().In(time.UTC)),
		Path:      req.GetPolicyContext().GetPath(),
		Policy: &api.DecisionPolicy{
			Context:        req.GetPolicyContext(),
			PolicyInstance: req.GetPolicyInstance(),
		},
		User: &api.DecisionUser{
			Context: req.GetIdentityContext(),
			Id:      getID(input),
			Email:   getEmail(input),
		},
		TenantId: getTenantID(ctx),
		Resource: req.GetResourceContext(),
		Outcomes: getOutcomes(resp.GetDecisions()),
	}

	if err := dlPlugin.Log(ctx, &d); err != nil {
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
		return item.GetDecision(), item.GetIs()
	})
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

func (*AuthorizerServer) isVerifyRequest(req *authorizer.IsRequest) error {
	if req.GetPolicyContext() == nil {
		return aerr.ErrInvalidArgument.Msg("policy context not set")
	}

	if req.GetPolicyContext().GetPath() == "" {
		return aerr.ErrInvalidArgument.Msg("policy context path not set")
	}

	if len(req.GetPolicyContext().GetDecisions()) == 0 {
		return aerr.ErrInvalidArgument.Msg("policy context decisions not set")
	}

	if req.GetResourceContext() == nil {
		req.ResourceContext = pb.NewStruct()
	}

	if req.GetIdentityContext() == nil {
		return aerr.ErrInvalidArgument.Msg("identity context not set")
	}

	if req.GetIdentityContext().GetType() == api.IdentityType_IDENTITY_TYPE_UNKNOWN {
		return aerr.ErrInvalidArgument.Msg("identity type UNKNOWN")
	}

	return nil
}

func (s *AuthorizerServer) isSetInput(ctx context.Context, req *authorizer.IsRequest) (map[string]any, error) {
	input := map[string]any{}

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
