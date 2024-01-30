package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/fatih/color"
	"github.com/pkg/errors"
)

type UninstallCmd struct {
	ContainerName    string `optional:"" default:"${container_name}" env:"CONTAINER_NAME" help:"container name"`
	ContainerVersion string `optional:"" default:"${container_version}" env:"CONTAINER_VERSION" help:"container version"`
}

func (cmd UninstallCmd) Run(c *cc.CommonCtx) error {
	color.Green(">>> uninstalling topaz...")

	if err := (StopCmd{}).Run(c); err != nil {
		return err
	}

	env := map[string]string{}

	args := []string{
		"images",
		cc.ContainerImage(
			cc.DefaultValue,      // service
			cc.DefaultValue,      // org
			cmd.ContainerName,    // name
			cmd.ContainerVersion, // version
		),
		"--filter", "label=org.opencontainers.image.source=https://github.com/aserto-dev/topaz",
		"-q",
	}

	str, err := dockerx.DockerWithOut(env, args...)
	if err != nil {
		return err
	}

	if str != "" {
		fmt.Fprintf(c.UI.Output(), "removing %s\n", "aserto-dev/topaz")
		if err := dockerx.DockerRun("rmi", str); err != nil {
			fmt.Fprintf(c.UI.Err(), "%s", err.Error())
		}
	}

	// remove config file last
	configFile := filepath.Join(cc.GetTopazCfgDir(), "config.yaml")
	if fi, err := os.Stat(configFile); err == nil && !fi.IsDir() {
		if err := os.Remove(configFile); err != nil {
			return errors.Wrap(err, "failed to delete topaz directory")
		}
	}

	return nil
}
