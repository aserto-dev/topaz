package authorizer

type AuthorizerCmd struct {
	CheckDecision CheckDecisionCmd `cmd:"" help:"evaluate policy decision" group:"authorizer"`
	DecisionTree  DecisionTreeCmd  `cmd:"" help:"get decision tree" group:"authorizer"`
	ExecQuery     ExecQueryCmd     `cmd:"" help:"execute query" group:"authorizer"`
	GetPolicy     GetPolicyCmd     `cmd:"" help:"get policy" group:"authorizer"`
	ListPolicies  ListPoliciesCmd  `cmd:"" help:"list policies" group:"authorizer"`
}
