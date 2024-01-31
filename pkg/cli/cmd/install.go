package cmd

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/fatih/color"
)

type InstallCmd struct {
	ContainerRegistry string `optional:"" default:"${container_registry}" env:"CONTAINER_REGISTRY" help:"container registry (host[:port]/repo)"`
	ContainerImage    string `optional:"" default:"${container_image}" env:"CONTAINER_IMAGE" help:"container image name"`
	ContainerTag      string `optional:"" default:"${container_tag}" env:"CONTAINER_TAG" help:"container tag"`
	ContainerPlatform string `optional:"" default:"${container_platform}" env:"CONTAINER_PLATFORM" help:"container platform"`
	ContainerVersion  string `optional:"" hidden:"" default:"" env:"CONTAINER_VERSION"`
}

func (cmd *InstallCmd) Run(c *cc.CommonCtx) error {
	cmd.ContainerTag = cc.ContainerVersionTag(cmd.ContainerVersion, cmd.ContainerTag)

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
		cc.Container(
			cmd.ContainerRegistry, // registry
			cmd.ContainerImage,    // image name
			cmd.ContainerTag,      // tag
		),
	}

	return dockerx.DockerWith(env, args...)
}
