package access

type AccessCmd struct {
	Evaluation     EvaluationCmd     `cmd:"" help:""`
	Evaluations    EvaluationsCmd    `cmd:"" help:""`
	SubjectSearch  SubjectSearchCmd  `cmd:"" help:""`
	ResourceSearch ResourceSearchCmd `cmd:"" help:""`
	ActionSearch   ActionSearchCmd   `cmd:"" help:""`
	Info           WellKnownCmd      `cmd:"" help:".well-known//authzen-configuration"`
}
