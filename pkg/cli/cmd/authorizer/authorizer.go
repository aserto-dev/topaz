package authorizer

type AuthorizerCmd struct {
	Test          TestCmd         `cmd:"" help:"execute authorizer assertions"`
	CheckDecision EvalCmd         `cmd:"" name:"eval" help:"evaluate policy decision"`
	DecisionTree  DecisionTreeCmd `cmd:"" name:"decisiontree" help:"get decision tree"`
	ExecQuery     QueryCmd        `cmd:"" name:"query" help:"execute query"`
	GetPolicy     GetPolicyCmd    `cmd:"" help:"get policy"`
	ListPolicies  ListPoliciesCmd `cmd:"" help:"list policies"`
}
