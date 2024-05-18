package authorizer

type AuthorizerCmd struct {
	CheckDecision EvalCmd         `cmd:"" name:"eval" help:"evaluate policy decision"`
	ExecQuery     QueryCmd        `cmd:"" name:"query" help:"execute query"`
	DecisionTree  DecisionTreeCmd `cmd:"" name:"decisiontree" help:"get decision tree"`
	GetPolicy     GetPolicyCmd    `cmd:"" help:"get policy"`
	ListPolicies  ListPoliciesCmd `cmd:"" help:"list policies"`
	Test          TestCmd         `cmd:"" help:"execute authorizer assertions"`
}
