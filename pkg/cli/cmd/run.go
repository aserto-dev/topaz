package cmd

import (
	"path/filepath"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"

	"github.com/fatih/color"
)

type RunCmd struct {
	StartRunCmd
}

func (cmd *RunCmd) Run(c *cc.CommonCtx) error {
	cmdX := cmd.StartRunCmd
	if cmd.ConfigFile != "" {
		c.Config.TopazConfigFile = filepath.Join(cc.GetTopazCfgDir(), cmd.ConfigFile)
		c.Config.ContainerName = cc.ContainerName(c.Config.TopazConfigFile)
		cmdX.ContainerName = c.Config.ContainerName
	}
	if c.CheckRunStatus(c.Config.ContainerName, cc.StatusRunning) {
		return ErrIsRunning
	}

	color.Green(">>> starting topaz...")

	args, err := cmdX.dockerArgs(c.Config.TopazConfigFile, true)
	if err != nil {
		return err
	}

	cmdArgs := []string{
		"run",
		"--config-file", "/config/" + filepath.Base(c.Config.TopazConfigFile),
	}

	args = append(args, cmdArgs...)

	return dockerx.DockerWith(cmdX.env(), args...)
}
