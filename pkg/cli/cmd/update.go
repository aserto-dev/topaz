package cmd

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/fatih/color"
)

type UpdateCmd struct {
	ContainerName     string `optional:"" default:"${container_name}" env:"CONTAINER_NAME" help:"container name"`
	ContainerVersion  string `optional:"" default:"${container_version}" env:"CONTAINER_VERSION" help:"container version"`
	ContainerPlatform string `optional:"" default:"${container_platform}" env:"CONTAINER_PLATFORM" help:"container platform"`
}

func (cmd UpdateCmd) Run(c *cc.CommonCtx) error {
	color.Green(">>> updating topaz...")

	env := map[string]string{}

	args := []string{
		"pull",
		"--platform", cmd.ContainerPlatform,
		cc.GetContainerImage(
			cc.DefaultValue,      // service
			cc.DefaultValue,      // org
			cmd.ContainerName,    // name
			cmd.ContainerVersion, // version
		),
	}

	return dockerx.DockerWith(env, args...)
}
