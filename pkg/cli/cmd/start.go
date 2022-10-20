package cmd

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"

	"github.com/fatih/color"
)

type StartCmd struct {
	ContainerName    string `optional:"" default:"topaz" help:"container name"`
	ContainerVersion string `optional:"" default:"latest" help:"container version" `
	Hostname         string `optional:"" help:"hostname for docker to set"`
}

func (cmd *StartCmd) Run(c *cc.CommonCtx) error {
	if running, err := dockerx.IsRunning(dockerx.Topaz); running || err != nil {
		if err != nil {
			return err
		}
		color.Yellow("!!! topaz is already running")
		return nil
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

var (
	dockerCmd = []string{
		"run",
	}

	dockerArgs = []string{
		"--rm",
		"--name", dockerx.Topaz,
		"--platform=linux/amd64",
		"-p", "8282:8282",
		"-p", "8383:8383",
		"-p", "8484:8484",
		"-v", "$TOPAZ_CERTS_DIR/certs:/certs:rw",
		"-v", "$TOPAZ_CFG_DIR/cfg:/config:ro",
		"-v", "$TOPAZ_EDS_DIR/db:/db:rw",
	}

	daemonArgs = []string{
		"-d",
	}

	containerName = []string{
		"ghcr.io/aserto-dev/$CONTAINER_NAME:$CONTAINER_VERSION",
	}

	hostname = []string{
		"--hostname", "$CONTAINER_HOSTNAME",
	}
)

func (cmd *StartCmd) dockerArgs() []string {
	args := append([]string{}, dockerCmd...)
	args = append(args, dockerArgs...)
	args = append(args, daemonArgs...)

	if cmd.Hostname != "" {
		args = append(args, hostname...)
	}

	return append(args, containerName...)
}

func (cmd *StartCmd) env(path string) map[string]string {
	return map[string]string{
		"TOPAZ_CERTS_DIR":    path,
		"TOPAZ_CFG_DIR":      path,
		"TOPAZ_EDS_DIR":      path,
		"CONTAINER_NAME":     cmd.ContainerName,
		"CONTAINER_VERSION":  cmd.ContainerVersion,
		"CONTAINER_HOSTNAME": cmd.Hostname,
	}
}
