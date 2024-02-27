package cmd

import (
	"os"
	"path/filepath"

	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/pkg/errors"

	"github.com/fatih/color"
)

type StartCmd struct {
	StartRunCmd
}

func (cmd *StartCmd) Run(c *cc.CommonCtx) error {
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

	args, err := cmdX.dockerArgs(c.Config.TopazConfigFile, false)
	if err != nil {
		return err
	}

	cmdArgs := []string{
		"run",
		"--config-file", "/config/" + filepath.Base(c.Config.TopazConfigFile),
	}

	args = append(args, cmdArgs...)

	if _, err := os.Stat(c.Config.TopazConfigFile); errors.Is(err, os.ErrNotExist) {
		return errors.Errorf("%s does not exist, please run 'topaz configure'", c.Config.TopazConfigFile)
	}

	generator := config.NewGenerator(filepath.Base(c.Config.TopazConfigFile))
	if _, err := generator.CreateCertsDir(); err != nil {
		return err
	}

	if _, err := generator.CreateDataDir(); err != nil {
		return err
	}

	return dockerx.DockerWith(cmdX.env(), args...)
}
