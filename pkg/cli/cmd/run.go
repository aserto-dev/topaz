package cmd

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"

	"github.com/fatih/color"
)

type RunCmd struct {
	ContainerName    string   `optional:"" default:"topaz" help:"container name"`
	ContainerVersion string   `optional:"" default:"latest" help:"container version" `
	Hostname         string   `optional:"" help:"hostname for docker to set"`
	Env              []string `optional:"" short:"e" help:"additional environment variable names to be passed to container"`
}

func (cmd *RunCmd) Run(c *cc.CommonCtx) error {
	if running, err := dockerx.IsRunning(dockerx.Topaz); running || err != nil {
		if !running {
			return ErrNotRunning
		}
		if err != nil {
			return err
		}
	}

	color.Green(">>> starting topaz...")

	args := cmd.dockerArgs()

	cmdArgs := []string{
		"run",
		"--config-file", "/config/config.yaml",
	}

	args = append(args, cmdArgs...)

	path, err := dockerx.DefaultRoots()
	if err != nil {
		return err
	}

	return dockerx.DockerWith(cmd.env(path), args...)
}

func (cmd *RunCmd) dockerArgs() []string {
	args := append([]string{}, dockerCmd...)
	args = append(args, "-ti")
	args = append(args, dockerArgs...)

	for _, env := range cmd.Env {
		args = append(args, "--env", env)
	}

	if cmd.Hostname != "" {
		args = append(args, hostname...)
	}

	return append(args, containerName...)
}

func (cmd *RunCmd) env(path string) map[string]string {
	return map[string]string{
		"TOPAZ_CERTS_DIR":    path,
		"TOPAZ_CFG_DIR":      path,
		"TOPAZ_EDS_DIR":      path,
		"CONTAINER_NAME":     cmd.ContainerName,
		"CONTAINER_VERSION":  cmd.ContainerVersion,
		"CONTAINER_HOSTNAME": cmd.Hostname,
	}
}
