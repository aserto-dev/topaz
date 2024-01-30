package cmd

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/fatih/color"
)

type InstallCmd struct {
	ContainerName     string `optional:"" default:"${container_name}" env:"CONTAINER_NAME" help:"container name"`
	ContainerVersion  string `optional:"" default:"${container_version}" env:"CONTAINER_VERSION" help:"container version"`
	ContainerPlatform string `optional:"" default:"${container_platform}" env:"CONTAINER_PLATFORM" help:"container platform"`
}

func (cmd InstallCmd) Run(c *cc.CommonCtx) error {
	if err := CheckRunning(c); err == nil {
		color.Yellow("!!! topaz is already running")
		return nil
	}

	color.Green(">>> installing topaz...")

	env := map[string]string{}

	args := []string{
		"pull",
		"--platform", cmd.ContainerPlatform,
		"--quiet",
		cc.ContainerImage(
			cc.DefaultValue,      // service
			cc.DefaultValue,      // org
			cmd.ContainerName,    // name
			cmd.ContainerVersion, // version
		),
	}

	return dockerx.DockerWith(env, args...)
}
