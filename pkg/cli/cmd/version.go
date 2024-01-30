package cmd

import (
	"fmt"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/aserto-dev/topaz/pkg/cli/x"
	"github.com/aserto-dev/topaz/pkg/version"
)

type VersionCmd struct {
	Container         bool   `short:"c" help:"display topazd container version" default:"false"`
	ContainerName     string `optional:"" default:"${container_name}" env:"CONTAINER_NAME" help:"container name"`
	ContainerVersion  string `optional:"" default:"${container_version}" env:"CONTAINER_VERSION" help:"container version"`
	ContainerPlatform string `optional:"" default:"${container_platform}" env:"CONTAINER_PLATFORM" help:"container platform"`
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
	// docker run -ti --rm --name topazd-version --platform linux/arm64 ghcr.io/aserto-dev/topaz:latest version
	args := []string{
		"run",
		"-ti",
		"--rm",
		"--name", "topazd-version",
		"--platform", cmd.ContainerPlatform,
		"--quiet",
		cc.ContainerImage(
			cc.DefaultValue,      // service
			cc.DefaultValue,      // org
			cmd.ContainerName,    // name
			cmd.ContainerVersion, // version
		),
		"version",
	}

	result, err := dockerx.DockerWithOut(env, args...)
	if err != nil {
		return err
	}
	fmt.Fprintf(c.UI.Output(), "%s\n", result)

	return nil
}
