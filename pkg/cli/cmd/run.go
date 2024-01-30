package cmd

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"

	"github.com/fatih/color"
)

type RunCmd struct {
	StartRunCmd
}

func (cmd *RunCmd) Run(c *cc.CommonCtx) error {
	if running, err := dockerx.IsRunning(cc.ContainerInstanceName()); running || err != nil {
		if !running {
			return ErrNotRunning
		}
		if err != nil {
			return err
		}
	}

	color.Green(">>> starting topaz...")
	cmdX := cmd.StartRunCmd
	args, err := cmdX.dockerArgs(cc.GetTopazDir(), true)
	if err != nil {
		return err
	}

	cmdArgs := []string{
		"run",
		"--config-file", "/config/config.yaml",
	}

	args = append(args, cmdArgs...)

	return dockerx.DockerWith(cmdX.env(), args...)
}
