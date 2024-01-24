package cmd

import (
	"fmt"
	"strings"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"

	"github.com/fatih/color"
)

type RunCmd struct {
	ContainerName     string   `optional:"" default:"${container_name}" help:"container name"`
	ContainerVersion  string   `optional:"" default:"${container_version}" help:"container version" `
	ContainerPlatform string   `optional:"" default:"${container_platform}" help:"container platform" `
	ContainerHostname string   `optional:"" name:"hostname" default:"${container_hostname}" help:"hostname for docker to set"`
	Env               []string `optional:"" short:"e" help:"additional environment variable names to be passed to container"`
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

	return dockerx.DockerWith(cmd.env(), args...)
}

func (cmd *RunCmd) dockerArgs(path string) ([]string, error) {
	args := append([]string{}, dockerCmd...)
	args = append(args, "-ti")

	policyRoot := dockerx.PolicyRoot()
	dockerArgs = append(dockerArgs, "-v", fmt.Sprintf("%s:/root/.policy:ro", policyRoot))
	args = append(args, dockerArgs...)

	volumes, err := getVolumes(path)
	if err != nil {
		return nil, err
	}
	args = append(args, volumes...)
	for i := range volumes {
		switch {
		case strings.Contains(volumes[i], "certs"):
			mountedPath := strings.Split(volumes[i], ":")[1]
			cmd.Env = append(cmd.Env, fmt.Sprintf("TOPAZ_CERTS_DIR=%s", mountedPath))
		case strings.Contains(volumes[i], "db"):
			mountedPath := strings.Split(volumes[i], ":")[1]
			cmd.Env = append(cmd.Env, fmt.Sprintf("TOPAZ_DB_DIR=%s", mountedPath))
		case strings.Contains(volumes[i], "cfg"):
			mountedPath := strings.Split(volumes[i], ":")[1]
			cmd.Env = append(cmd.Env, fmt.Sprintf("TOPAZ_CFG_DIR=%s", mountedPath))
		}
	}

	ports, err := getPorts(path)
	if err != nil {
		return nil, err
	}
	args = append(args, ports...)

	for _, env := range cmd.Env {
		args = append(args, "--env", env)
	}

	if cmd.ContainerHostname != "" {
		args = append(args, hostname...)
	}

	return append(args, containerName...), nil
}

func (cmd *RunCmd) env() map[string]string {
	return map[string]string{
		"TOPAZ_CERTS_DIR":    cc.GetTopazCertsDir(),
		"TOPAZ_CFG_DIR":      cc.GetTopazCfgDir(),
		"TOPAZ_DB_DIR":       cc.GetTopazDataDir(),
		"CONTAINER_NAME":     cc.GetContainerName(),
		"CONTAINER_VERSION":  cc.GetContainerVersion(),
		"CONTAINER_PLATFORM": cc.GetContainerPlatform(),
		"CONTAINER_HOSTNAME": cc.GetContainerHostname(),
	}
}
