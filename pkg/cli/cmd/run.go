package cmd

import (
	"errors"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"

	"github.com/fatih/color"
)

type RunCmd struct {
	StartRunCmd
}

func (cmd *RunCmd) Run(c *cc.CommonCtx) error {
	err := checkRunning(c, cmd.ContainerName)
	if err != nil && !errors.Is(err, ErrNotRunning) {
		return err
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
