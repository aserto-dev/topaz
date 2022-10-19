package cmd

import (
	"fmt"
	"os"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/fatih/color"
	"github.com/pkg/errors"
)

type UninstallCmd struct{}

func (cmd UninstallCmd) Run(c *cc.CommonCtx) error {
	color.Green(">>> uninstalling topaz...")

	var err error

	//nolint :gocritic // tbd
	if err = (StopCmd{}).Run(c); err != nil {
		return err
	}

	path, err := dockerx.DefaultRoots()
	if err != nil {
		return err
	}

	if err = os.RemoveAll(path); err != nil {
		return errors.Wrap(err, "failed to delete topaz directory")
	}

	str, err := dockerx.DockerWithOut(map[string]string{
		"NAME": "topaz",
	},
		"images",
		"ghcr.io/aserto-dev/$NAME",
		"--filter", "label=org.opencontainers.image.source=https://github.com/aserto-dev/topaz",
		"-q",
	)
	if err != nil {
		return err
	}

	if str != "" {
		fmt.Fprintf(c.UI.Output(), "removing %s\n", "aserto-dev/topaz")
		err = dockerx.DockerRun("rmi", str)
	}

	return err
}
