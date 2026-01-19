package impl

import (
	"context"
	"fmt"
	"strings"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	runtime "github.com/aserto-dev/runtime"

	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *AuthorizerServer) DecisionTree(ctx context.Context, req *authorizer.DecisionTreeRequest) (*authorizer.DecisionTreeResponse, error) {
	log := s.logger.With().Str("api", "decision_tree").Logger()

	if err := s.decisionTreeVerifyRequest(req); err != nil {
		return &authorizer.DecisionTreeResponse{}, err
	}

	input, err := s.decisionTreeSetInput(ctx, req)
	if err != nil {
		return &authorizer.DecisionTreeResponse{}, err
	}

	log.Debug().Interface("input", input).Msg("decision_tree")

	policyRuntime, err := s.getRuntime(ctx, req.GetPolicyInstance())
	if err != nil {
		return &authorizer.DecisionTreeResponse{}, err
	}

	queryStmt, err := s.decisionTreeBuildQuery(ctx, policyRuntime, req)
	if err != nil {
		return &authorizer.DecisionTreeResponse{}, err
	}

	query, err := rego.New(
		rego.Compiler(policyRuntime.GetPluginsManager().GetCompiler()),
		rego.Store(policyRuntime.GetPluginsManager().Store),
		rego.Query(queryStmt.String()),
	).PrepareForEval(ctx)
	if err != nil {
		return &authorizer.DecisionTreeResponse{}, aerr.ErrBadQuery.Err(err).Msg(queryStmt.String())
	}

	queryResults, err := query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return &authorizer.DecisionTreeResponse{}, aerr.ErrBadQuery.Err(err).Msgf("query evaluation failed: %s", queryStmt.String())
	}

	if len(queryResults) == 0 {
		return &authorizer.DecisionTreeResponse{}, aerr.ErrBadQuery.Err(err).Msgf("undefined results: %s", queryStmt.String())
	}

	resultBuilder := make(map[string]any)

	for _, expression := range queryResults[0].Expressions {
		expr := strings.Split(expression.Text, "=")
		rule := strings.TrimSpace(expr[0])
		path := strings.TrimPrefix(strings.TrimSpace(expr[1]), "data.")

		if req.GetOptions().GetPathSeparator() == authorizer.PathSeparator_PATH_SEPARATOR_SLASH {
			path = strings.ReplaceAll(path, ".", "/")
		}

		binding, ok := queryResults[0].Bindings[rule].(map[string]any)
		if !ok {
			continue
		}

		outcomes := make(map[string]any)

		for _, decision := range req.GetPolicyContext().GetDecisions() {
			if r, ok := binding[decision].(bool); ok {
				outcomes[decision] = r
			}
		}

		resultBuilder[path] = outcomes
	}

	paths, err := structpb.NewStruct(resultBuilder)
	if err != nil {
		return &authorizer.DecisionTreeResponse{}, err
	}

	resp := &authorizer.DecisionTreeResponse{
		PathRoot: req.GetPolicyContext().GetPath(),
		Path:     paths,
	}

	return resp, nil
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

func (*AuthorizerServer) decisionTreeVerifyRequest(req *authorizer.DecisionTreeRequest) error {
	if req.GetPolicyContext() == nil {
		return aerr.ErrInvalidArgument.Msg("policy context not set")
	}

	if req.GetPolicyContext().GetPath() == "" {
		return aerr.ErrInvalidArgument.Msg("policy context path not set")
	}

	if req.GetIdentityContext() == nil || req.GetIdentityContext().GetType() == api.IdentityType_IDENTITY_TYPE_UNKNOWN {
		return aerr.ErrInvalidArgument.Msg("identity type UNKNOWN")
	}

	if req.GetResourceContext() == nil {
		req.ResourceContext = &structpb.Struct{Fields: make(map[string]*structpb.Value)}
	}

	if req.GetOptions() == nil {
		req.Options = &authorizer.DecisionTreeOptions{
			PathSeparator: authorizer.PathSeparator_PATH_SEPARATOR_DOT,
		}
	}

	return nil
}

func (s *AuthorizerServer) decisionTreeSetInput(ctx context.Context, req *authorizer.DecisionTreeRequest) (map[string]any, error) {
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

func (*AuthorizerServer) decisionTreeBuildQuery(
	ctx context.Context,
	runtime *runtime.Runtime,
	req *authorizer.DecisionTreeRequest,
) (
	strings.Builder,
	error,
) {
	listPolicies, err := runtime.ListPolicies(ctx)
	if err != nil {
		return strings.Builder{}, errors.Wrap(err, "get policy list")
	}

	decisionFilter := initDecisionFilter(req.GetPolicyContext().GetDecisions())

	queryStmt := strings.Builder{}
	r := 0

	for i, policy := range listPolicies {
		for _, rule := range policy.AST.Rules {
			if !strings.HasPrefix(policy.AST.Package.Path.String(), "data."+req.GetPolicyContext().GetPath()) {
				continue
			}

			if !decisionFilter(rule.Head.Name.String()) {
				continue
			}

			queryStmt.WriteString(fmt.Sprintf("r%d = %s\n", i, policy.AST.Package.Path))

			r++

			break
		}
	}

	if queryStmt.Len() == 0 {
		return strings.Builder{}, aerr.ErrInvalidArgument.Msg("no decisions specified")
	}

	return queryStmt, nil
}
