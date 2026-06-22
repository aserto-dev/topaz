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
	decisionlog_plugin "github.com/aserto-dev/topaz/topazd/authorizer/plugins/decisionlog"

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

	rt, err := s.getRuntime(ctx)
	if err != nil {
		return &authorizer.IsResponse{}, err
	}

	// The Rego query body and its prepared form depend only on the policy
	// path and the decisions list — both stable for the lifetime of the
	// active OPA compiler. Cache the PreparedEvalQuery so repeated Is()
	// calls for the same (path, decisions) skip the parse + plan work and
	// stop fighting each other on the compiler's internal locks. The cache
	// is invalidated whenever the compiler is rotated (bundle reload).
	policyPath := req.GetPolicyContext().GetPath()
	decisions := req.GetPolicyContext().GetDecisions()
	preparedKey := cacheKey(policyPath, decisions)

	query, err := s.preparedQueries.getOrPrepare(ctx, rt, preparedKey, func(ctx context.Context) (rego.PreparedEvalQuery, error) {
		queryStmt := strings.Builder{}

		for i, decision := range decisions {
			rule := fmt.Sprintf("data.%s.%s\n", policyPath, decision)

			if ok, err := rt.ValidateRule(rule); !ok {
				return rego.PreparedEvalQuery{}, aerr.ErrBadQuery.Err(err).Msgf("invalid rule: %q", rule)
			}

			q := fmt.Sprintf("x%d = %s\n", i, rule)

			if _, err := rt.ValidateQuery(q); err != nil {
				return rego.PreparedEvalQuery{}, aerr.ErrBadQuery.Err(err).Msgf("invalid query: %q", q)
			}

			queryStmt.WriteString(q)
		}

		pq, err := rt.ValidateQuery(queryStmt.String())
		if err != nil {
			return rego.PreparedEvalQuery{}, aerr.ErrBadQuery.Err(err).Msgf("invalid query batch: %q", queryStmt.String())
		}

		prepared, err := rego.New(
			rego.Compiler(rt.GetPluginsManager().GetCompiler()),
			rego.Store(rt.GetPluginsManager().Store),
			rego.ParsedQuery(pq),
		).PrepareForEval(ctx)
		if err != nil {
			return rego.PreparedEvalQuery{}, aerr.ErrBadQuery.Err(err).Msg(queryStmt.String())
		}

		return prepared, nil
	})
	if err != nil {
		return &authorizer.IsResponse{}, err
	}

	resp := &authorizer.IsResponse{
		Decisions: []*authorizer.Decision{},
	}

	queryResults, err := query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return resp, aerr.ErrBadQuery.Err(err).Msgf("query evaluation failed: path=%s decisions=%v", policyPath, decisions)
	}

	if len(queryResults) == 0 {
		return resp, aerr.ErrBadQuery.Err(err).Msgf("undefined results: path=%s decisions=%v", policyPath, decisions)
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

	dlPlugin := decisionlog_plugin.Lookup(rt.GetPluginsManager())
	if dlPlugin == nil {
		return resp, err
	}

	d := api.Decision{
		Id:        uuid.NewString(),
		Timestamp: timestamppb.New(time.Now().In(time.UTC)),
		Path:      req.GetPolicyContext().GetPath(),
		Policy: &api.DecisionPolicy{
			Context: req.GetPolicyContext(),
		},
		User: &api.DecisionUser{
			Context: req.GetIdentityContext(),
			Id:      getID(input),
			Email:   getEmail(input),
		},
		Resource: req.GetResourceContext(),
		Outcomes: getOutcomes(resp.GetDecisions()),
	}

	if err := dlPlugin.Log(ctx, &d); err != nil {
		return resp, err
	}

	return resp, err
}

func getOutcomes(decisions []*authorizer.Decision) map[string]bool {
	return lo.SliceToMap(decisions, func(item *authorizer.Decision) (string, bool) {
		return item.GetDecision(), item.GetIs()
	})
}

func getID(v map[string]any) string {
	if u, ok := v["user"].(map[string]any); ok {
		if i, ok := u["id"].(string); ok {
			return i
		}
	}

	return ""
}

func getEmail(v map[string]any) string {
	if u, ok := v["user"].(map[string]any); ok {
		if e, ok := u["email"].(string); ok {
			return e
		}

		if p, ok := u["properties"].(map[string]any); ok {
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
