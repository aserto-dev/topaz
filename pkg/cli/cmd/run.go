package cmd

import (
	"fmt"

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

	rootPath, err := dockerx.DefaultRoots()
	if err != nil {
		return err
	}

	color.Green(">>> starting topaz...")

	args, err := cmd.dockerArgs(rootPath)
	if err != nil {
		return err
	}

	cmdArgs := []string{
		"run",
		"--config-file", "/config/config.yaml",
	}

	args = append(args, cmdArgs...)

	return dockerx.DockerWith(cmd.env(rootPath), args...)
}

func (cmd *RunCmd) dockerArgs(path string) ([]string, error) {
	args := append([]string{}, dockerCmd...)
	args = append(args, "-ti")

	policyRoot := dockerx.PolicyRoot()
	dockerArgs = append(dockerArgs, "-v", fmt.Sprintf("%s:/root/.policy:ro", policyRoot))
	args = append(args, dockerArgs...)

	for _, env := range cmd.Env {
		args = append(args, "--env", env)
	}

	volumes, err := getVolumes(path)
	if err != nil {
		return nil, err
	}
	args = append(args, volumes...)

	ports, err := getPorts(path)
	if err != nil {
		return nil, err
	}
	args = append(args, ports...)

	if cmd.Hostname != "" {
		args = append(args, hostname...)
	}

	return append(args, containerName...), nil
}

func (cmd *RunCmd) env(path string) map[string]string {
	return map[string]string{
		"TOPAZ_CERTS_DIR":    path,
		"TOPAZ_CFG_DIR":      path,
		"TOPAZ_DB_DIR":       path,
		"CONTAINER_NAME":     cmd.ContainerName,
		"CONTAINER_VERSION":  cmd.ContainerVersion,
		"CONTAINER_HOSTNAME": cmd.Hostname,
	}
}
