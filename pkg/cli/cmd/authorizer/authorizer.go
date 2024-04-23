package authorizer

type AuthorizerCmd struct {
	Test TestCmd `cmd:"" help:"execute authorizer assertions"`
}
