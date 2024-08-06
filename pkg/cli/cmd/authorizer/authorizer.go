package authorizer

type AuthorizerCmd struct {
	CheckDecision EvalCmd         `cmd:"" name:"eval" help:"evaluate policy decision"`
	ExecQuery     QueryCmd        `cmd:"" name:"query" help:"execute query"`
	DecisionTree  DecisionTreeCmd `cmd:"" name:"decisiontree" help:"get decision tree"`
	Get           GetCmd          `cmd:"" help:"get policy"`
	List          ListCmd         `cmd:"" help:"list policy"`
	Test          TestCmd         `cmd:"" help:"execute authorizer assertions"`
}

type GetCmd struct {
	Policy GetPolicyCmd `cmd:"" help:"get policy"`
}

type ListCmd struct {
	Policies ListPoliciesCmd `cmd:"" help:"list policies"`
}
