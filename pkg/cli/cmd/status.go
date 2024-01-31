package cmd

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/fatih/color"
)

type StatusCmd struct{}

func (cmd StatusCmd) Run(c *cc.CommonCtx) error {
	running, err := dockerx.IsRunning(cc.ContainerName())
	if err != nil {
		return err
	}
	if running {
		color.Green(">>> topaz is running")
	} else {
		color.Yellow(">>> topaz is not running")
	}
	return nil
}
