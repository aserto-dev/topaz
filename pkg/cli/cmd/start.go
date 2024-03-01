package cmd

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
)

type StartCmd struct {
	StartRunCmd
}

func (cmd *StartCmd) Run(c *cc.CommonCtx) error {
	return cmd.run(c, modeDaemon)
}
