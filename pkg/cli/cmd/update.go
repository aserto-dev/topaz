package cmd

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/fatih/color"
)

type UpdateCmd struct {
	ContainerName     string `optional:"" default:"${container_name}" help:"container name"`
	ContainerVersion  string `optional:"" default:"${container_version}" help:"container version" `
	ContainerPlatform string `optional:"" default:"${container_platform}" help:"container platform" `
	ContainerHostname string `optional:"" name:"hostname" default:"${container_hostname}" help:"hostname for docker to set"`
}

func (cmd UpdateCmd) Run(c *cc.CommonCtx) error {
	color.Green(">>> updating topaz...")

	args := []string{}
	args = append(args, "pull")
	args = append(args, platform...)
	if cmd.ContainerHostname != "" {
		args = append(args, hostname...)
	}
	args = append(args, containerName...)

	return dockerx.DockerWith(map[string]string{
		"CONTAINER_NAME":     cmd.ContainerName,
		"CONTAINER_VERSION":  cmd.ContainerVersion,
		"CONTAINER_PLATFORM": cmd.ContainerPlatform,
	},
		args...,
	)
}
