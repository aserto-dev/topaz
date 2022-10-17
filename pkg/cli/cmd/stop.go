package cmd

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/fatih/color"
)

type StopCmd struct{}

func (cmd StopCmd) Run(c *cc.CommonCtx) error {
	running, err := dockerx.IsRunning(dockerx.Topaz)
	if err != nil {
		return err
	}

	if running {
		color.Green(">>> stopping topaz...")
		return dockerx.DockerRun("stop", dockerx.Topaz)
	}

	return nil
}
