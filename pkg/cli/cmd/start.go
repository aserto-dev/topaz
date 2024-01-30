package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

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
	if cmd.ConfigFile != "" {
		c.Config.DefaultConfigFile = cmd.ConfigFile
	}
	if err := CheckRunning(c); err == nil {
		return fmt.Errorf("topaz is already running")
	}

	color.Green(">>> starting topaz...")

	args, err := cmd.dockerArgs(c.Config.DefaultConfigFile)
	if err != nil {
		return err
	}

	cmdArgs := []string{
		"run",
		"--config-file", "/config/" + c.Config.DefaultConfigFile,
	}

	args = append(args, cmdArgs...)

	if _, err := os.Stat(path.Join(cc.GetTopazCfgDir(), c.Config.DefaultConfigFile)); errors.Is(err, os.ErrNotExist) {
		return errors.Errorf("%s does not exist, please run 'topaz configure'", path.Join(cc.GetTopazCfgDir(), c.Config.DefaultConfigFile))
	}

	generator := config.NewGenerator(c.Config.DefaultConfigFile)
	if _, err := generator.CreateCertsDir(); err != nil {
		return err
	}

	if _, err := generator.CreateDataDir(); err != nil {
		return err
	}

	return dockerx.DockerWith(cmdX.env(), args...)
}
