package cmd

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/fatih/color"
)

type InstallCmd struct {
	ContainerName    string `optional:""  default:"topaz" help:"container name"`
	ContainerVersion string `optional:""  default:"latest" help:"container version"`
}

func (cmd InstallCmd) Run(c *cc.CommonCtx) error {
	if err := CheckRunning(c); err == nil {
		color.Yellow("!!! topaz is already running")
		return nil
	}

	color.Green(">>> installing topaz...")

	return dockerx.DockerWith(map[string]string{
		"CONTAINER_NAME":    cmd.ContainerName,
		"CONTAINER_VERSION": cmd.ContainerVersion,
	},
		"pull", "ghcr.io/aserto-dev/$CONTAINER_NAME:$CONTAINER_VERSION",
	)
}
