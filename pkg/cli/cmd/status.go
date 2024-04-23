package cmd

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/fatih/color"
)

type StatusCmd struct {
	ContainerName string `optional:"" default:"${container_name}" env:"CONTAINER_NAME" help:"container name"`
}

func (cmd *StatusCmd) Run(c *cc.CommonCtx) error {
	containerName := c.Config.Running.ContainerName
	if cmd.ContainerName != containerName {
		containerName = cmd.ContainerName
	}

	if c.CheckRunStatus(containerName, cc.StatusRunning) {
		color.Green(">>> topaz %q is running", c.Config.Running.Config)
	} else {
		color.Yellow(">>> topaz is not running")
	}

	return nil
}
