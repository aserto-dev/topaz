package cmd

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/fatih/color"
)

type StatusCmd struct {
	ContainerName string `optional:"" default:"${container_name}" env:"CONTAINER_NAME" help:"container name"`
}

func (cmd *StatusCmd) Run(c *cc.CommonCtx) error {
	if c.CheckRunStatus(cmd.ContainerName, cc.StatusRunning) {
		color.Green(">>> topaz is running")
	} else {
		color.Yellow(">>> topaz is not running")
	} else {
		color.Green(">>> topaz is running")
	}
	return nil
}
