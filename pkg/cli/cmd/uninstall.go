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
	ContainerRegistry string `optional:"" default:"${container_registry}" env:"CONTAINER_REGISTRY" help:"container registry (host[:port]/repo)"`
	ContainerImage    string `optional:"" default:"${container_image}" env:"CONTAINER_IMAGE" help:"container image name"`
	ContainerTag      string `optional:"" default:"${container_tag}" env:"CONTAINER_TAG" help:"container tag"`
	ContainerName     string `optional:"" default:"${container_name}" env:"CONTAINER_NAME" help:"container name"`
}

func (cmd UninstallCmd) Run(c *cc.CommonCtx) error {
	color.Green(">>> uninstalling topaz...")

	if err := (StopCmd{ContainerName: cmd.ContainerName}).Run(c); err != nil {
		return err
	}

	env := map[string]string{}

	args := []string{
		"images",
		cc.Container(
			cmd.ContainerRegistry, // registry
			cmd.ContainerImage,    // image
			cmd.ContainerTag,      // tag
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

	// remove directory.db file.
	if err := removeFile(filepath.Join(cc.GetTopazDataDir(), "directory.db")); err != nil {
		return err
	}
	// remove directory.db.sync file.
	if err := removeFile(filepath.Join(cc.GetTopazDataDir(), "directory.db.sync")); err != nil {
		return err
	}
	// remove config.yaml file last.
	if err := removeFile(filepath.Join(cc.GetTopazCfgDir(), "config.yaml")); err != nil {
		return err
	}

	return nil
}

func removeFile(fpath string) error {
	if fi, err := os.Stat(fpath); err == nil && !fi.IsDir() {
		if err := os.Remove(fpath); err != nil {
			return errors.Wrapf(err, "failed to delete %s", fpath)
		}
	}
	return nil
}
