package cmd

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
)

type RunCmd struct {
	*dockerx.Container `embed:""`
}

func (cmd *RunCmd) Run(c *cc.CommonCtx) error {
	if err := createMountDirs(); err != nil {
		return err
	}
	return cmd.Start(dockerx.Interactive)
}
