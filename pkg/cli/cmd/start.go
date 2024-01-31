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
		c.Config.DefaultConfigFile = filepath.Join(cc.GetTopazCfgDir(), cmd.ConfigFile)
		c.Config.ContainerName = cc.ContainerName(c.Config.DefaultConfigFile)
	}
	if err := CheckRunning(c); err == nil {
		return ErrIsRunning
	}
	color.Green(">>> starting topaz...")

	args, err := cmdX.dockerArgs(c.Config.DefaultConfigFile, false)
	if err != nil {
		return err
	}

	cmdArgs := []string{
		"run",
		"--config-file", "/config/" + filepath.Base(c.Config.DefaultConfigFile),
	}

	args = append(args, cmdArgs...)

	if _, err := os.Stat(c.Config.DefaultConfigFile); errors.Is(err, os.ErrNotExist) {
		return errors.Errorf("%s does not exist, please run 'topaz configure'", c.Config.DefaultConfigFile)
	}

	generator := config.NewGenerator(filepath.Base(c.Config.DefaultConfigFile))
	if _, err := generator.CreateCertsDir(); err != nil {
		return err
	}

	if _, err := generator.CreateDataDir(); err != nil {
		return err
	}

	return dockerx.DockerWith(cmdX.env(), args...)
}
