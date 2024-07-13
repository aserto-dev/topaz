package topaz

import (
	"fmt"

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
		fmt.Fprint(c.StdOut(), color.GreenString(">>> topaz %q is running\n", c.Config.Running.Config))
	} else {
		fmt.Fprint(c.StdOut(), color.YellowString(">>> topaz is not running\n"))
	}

	return nil
}
