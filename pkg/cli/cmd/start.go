package cmd

import (
	"os"
	"path"

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

	if c.CheckRunStatus(cmd.ContainerName, cc.StatusRunning) {
		return ErrIsRunning
	}

	color.Green(">>> starting topaz...")
	args, err := cmdX.dockerArgs(cc.GetTopazDir(), false)
	if err != nil {
		return err
	}

	cmdArgs := []string{
		"run",
		"--config-file", "/config/config.yaml",
	}

	args = append(args, cmdArgs...)

	if _, err := os.Stat(path.Join(cc.GetTopazCfgDir(), "config.yaml")); errors.Is(err, os.ErrNotExist) {
		return errors.Errorf("%s does not exist, please run 'topaz configure'", path.Join(cc.GetTopazCfgDir(), "config.yaml"))
	}

	generator := config.NewGenerator("config.yaml")
	if _, err := generator.CreateCertsDir(); err != nil {
		return err
	}

	if _, err := generator.CreateDataDir(); err != nil {
		return err
	}

	return dockerx.DockerWith(cmdX.env(), args...)
}
