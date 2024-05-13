package topaz

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
)

type RunCmd struct {
	StartRunCmd
}

func (cmd *RunCmd) Run(c *cc.CommonCtx) error {
	return cmd.run(c, modeInteractive)
}
