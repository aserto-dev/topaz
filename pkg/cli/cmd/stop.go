package cmd

import (
	"fmt"
	"strings"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/fatih/color"
)

type StopCmd struct {
	ContainerName string `optional:"" default:"${container_name}" env:"CONTAINER_NAME" help:"container name"`
}

func (cmd *StopCmd) Run(c *cc.CommonCtx) error {
	if c.CheckRunStatus(cmd.ContainerName, cc.StatusNotRunning) {
		return nil
	}

	color.Green(">>> stopping topaz...")
	return dockerx.DockerRun("stop", cmd.ContainerName)
}
