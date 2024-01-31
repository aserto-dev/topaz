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
	if cmd.ConfigFile != "" {
		c.Config.DefaultConfigFile = filepath.Join(cc.GetTopazCfgDir(), cmd.ConfigFile)
		c.Config.ContainerName = cc.ContainerName(c.Config.DefaultConfigFile)
	}
	if c.CheckRunStatus(cmd.ContainerName, cc.StatusRunning) {
		return ErrIsRunning
	}

	color.Green(">>> starting topaz...")
	cmdX := cmd.StartRunCmd
	args, err := cmdX.dockerArgs(c.Config.DefaultConfigFile, true)
	if err != nil {
		return err
	}

	cmdArgs := []string{
		"run",
		"--config-file", "/config/" + filepath.Base(c.Config.DefaultConfigFile),
	}

	args = append(args, cmdArgs...)

	return dockerx.DockerWith(cmdX.env(), args...)
}
