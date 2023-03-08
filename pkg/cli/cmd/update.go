package cmd

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/fatih/color"
)

type UpdateCmd struct {
	ContainerName    string `optional:"" default:"topaz" help:"container name"`
	ContainerVersion string `optional:"" default:"latest" help:"container version" `
	Hostname         string `optional:"" help:"hostname for docker to set"`
}

func (cmd UpdateCmd) Run(c *cc.CommonCtx) error {
	color.Green(">>> updating topaz...")

	args := []string{}
	args = append(args, "pull")
	args = append(args, dockerx.Platform...)
	if cmd.Hostname != "" {
		args = append(args, dockerx.Hostname...)
	}
	args = append(args, dockerx.ImageName...)

	return dockerx.DockerWith(map[string]string{
		"CONTAINER_NAME":    cmd.ContainerName,
		"CONTAINER_VERSION": cmd.ContainerVersion,
	},
		args...,
	)
}
