package cmd

import (
	"fmt"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/aserto-dev/topaz/pkg/cli/x"
	"github.com/aserto-dev/topaz/pkg/version"
)

type VersionCmd struct {
	Container        bool   `short:"c" help:"display topazd container version" default:"false"`
	ContainerName    string `optional:"" default:"topaz" help:"container name"`
	ContainerVersion string `optional:"" default:"latest" help:"container version"`
	Platform         string `optional:"" default:"linux/amd64" help:"set platform if server is multi-platform capable"`
}

func (cmd *VersionCmd) Run(c *cc.CommonCtx) error {
	fmt.Fprintf(c.UI.Output(), "%s %s\n",
		x.AppName,
		version.GetInfo().String(),
	)

	if !cmd.Container {
		return nil
	}

	env := map[string]string{}

	// default command
	// docker run -ti --rm --name topazd-version --platform=linux/arm64 ghcr.io/aserto-dev/topaz:latest version
	args := []string{
		"run",
		"-ti",
		"--rm",
		"--name", "topazd-version",
		"--platform=" + cmd.Platform,
		"ghcr.io/aserto-dev/" + cmd.ContainerName + ":" + cmd.ContainerVersion,
		"version",
	}

	result, err := dockerx.DockerWithOut(env, args...)
	if err != nil {
		return err
	}
	fmt.Fprintf(c.UI.Output(), "%s\n", result)

	return nil
}
