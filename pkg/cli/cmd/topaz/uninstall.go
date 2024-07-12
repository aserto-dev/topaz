package topaz

import (
	"os"
	"path/filepath"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/pkg/errors"
)

type UninstallCmd struct {
	ContainerRegistry string `optional:"" default:"${container_registry}" env:"CONTAINER_REGISTRY" help:"container registry (host[:port]/repo)"`
	ContainerImage    string `optional:"" default:"${container_image}" env:"CONTAINER_IMAGE" help:"container image name"`
	ContainerTag      string `optional:"" default:"${container_tag}" env:"CONTAINER_TAG" help:"container tag"`
	ContainerName     string `optional:"" default:"${container_name}" env:"CONTAINER_NAME" help:"container name"`
	ContainerVersion  string `optional:"" hidden:"" default:"" env:"CONTAINER_VERSION"`
}

func (cmd *UninstallCmd) Run(c *cc.CommonCtx) error {
	// stop container instance, if running.
	if err := (&StopCmd{ContainerName: cmd.ContainerName, Wait: true}).Run(c); err != nil {
		return err
	}

	c.Con().Info().Msg(">>> uninstalling %s...",
		cc.Container(
			cmd.ContainerRegistry, // registry
			cmd.ContainerImage,    // image name
			cmd.ContainerTag,      // tag
		),
	)

	dc, err := dockerx.New()
	if err != nil {
		return err
	}

	// remove container image when exists.
	if err := dc.RemoveImage(cc.Container(
		cmd.ContainerRegistry,
		cmd.ContainerImage,
		cc.ContainerVersionTag(cmd.ContainerVersion, cmd.ContainerTag),
	)); err != nil {
		return err
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
