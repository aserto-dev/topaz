package topaz

import (
	"context"

	"github.com/aserto-dev/topaz/topaz/cc"
)

type RunCmd struct {
	StartRunCmd
}

func (cmd *RunCmd) Run(ctx context.Context, cfg *cc.Config) error {
	return cmd.run(cfg, modeInteractive)
}
