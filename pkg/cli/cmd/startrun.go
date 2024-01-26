package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/aserto-dev/topaz/pkg/cli/g"
)

type StartRunCmd struct {
	ContainerName     string   `optional:"" default:"${container_name}" env:"CONTAINER_VERSION" help:"container name"`
	ContainerVersion  string   `optional:"" default:"${container_version}" env:"CONTAINER_VERSION" help:"container version" `
	ContainerPlatform string   `optional:"" default:"${container_platform}" env:"CONTAINER_PLATFORM" help:"container platform" `
	ContainerHostname string   `optional:"" name:"hostname" default:"" env:"CONTAINER_HOSTNAME" help:"hostname for docker to set"`
	Env               []string `optional:"" short:"e" help:"additional environment variable names to be passed to container"`
}

func (cmd *StartRunCmd) dockerArgs(rootPath string, interactive bool) ([]string, error) {
	args := []string{
		"run",
		"--rm",
		"--name", dockerx.Topaz,
		g.Iff(interactive, "-ti", "-d"),
	}

	policyRoot := dockerx.PolicyRoot()
	args = append(args, "-v", fmt.Sprintf("%s:/root/.policy:ro", policyRoot))

	volumes, err := getVolumes(rootPath)
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

	ports, err := getPorts(rootPath)
	if err != nil {
		return nil, err
	}
	args = append(args, ports...)

	for _, env := range cmd.Env {
		args = append(args, "--env", env)
	}

	if cmd.ContainerHostname != "" {
		args = append(args, "--hostname", cmd.ContainerHostname)
	}

	return append(args,
		cc.GetContainerImage(
			cc.DefaultValue,          // service
			cc.DefaultValue,          // org
			cc.GetContainerName(),    // name
			cc.GetContainerVersion(), // version
		),
	), nil
}

func (cmd *StartRunCmd) env() map[string]string {
	return map[string]string{
		"TOPAZ_DIR":       cc.GetTopazDir(),
		"TOPAZ_CERTS_DIR": cc.GetTopazCertsDir(),
		"TOPAZ_CFG_DIR":   cc.GetTopazCfgDir(),
		"TOPAZ_DB_DIR":    cc.GetTopazDataDir(),
	}
}

func getPorts(rootPath string) ([]string, error) {
	portMap := make(map[string]string)
	configLoader, err := config.LoadConfiguration(fmt.Sprintf("%s/cfg/config.yaml", rootPath))
	if err != nil {
		return nil, err
	}

	portArray, err := configLoader.GetPorts()
	if err != nil {
		return nil, err
	}

	for i := range portArray {
		portMap[portArray[i]] = fmt.Sprintf("%s:%s", portArray[i], portArray[i])
	}

	// ensure unique assignment for each port
	var args []string
	for _, v := range portMap {
		args = append(args, "-p", v)
	}
	return args, nil
}

func getVolumes(rootPath string) ([]string, error) {
	volumeMap := make(map[string]string)
	configPath := fmt.Sprintf("%s/cfg/config.yaml", rootPath)
	configLoader, err := config.LoadConfiguration(configPath)
	if err != nil {
		return nil, err
	}

	paths, err := configLoader.GetPaths()
	if err != nil {
		return nil, err
	}

	for i := range paths {
		directory := filepath.Dir(paths[i])
		volumeMap[directory] = fmt.Sprintf("%s:%s", directory, fmt.Sprintf("/%s", filepath.Base(directory)))
	}

	// manually attach the configuration folder
	args := []string{"-v", fmt.Sprintf("%s:/config:ro", filepath.Dir(configPath))}
	for _, v := range volumeMap {
		args = append(args, "-v", v)
	}
	return args, nil
}
