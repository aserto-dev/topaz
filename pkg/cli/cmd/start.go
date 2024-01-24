package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/pkg/errors"

	"github.com/fatih/color"
)

type StartCmd struct {
	ContainerName     string   `optional:"" default:"${container_name}" help:"container name"`
	ContainerVersion  string   `optional:"" default:"${container_version}" help:"container version" `
	ContainerPlatform string   `optional:"" default:"${container_platform}" help:"container platform"`
	ContainerHostname string   `optional:"" name:"hostname" default:"${container_hostname}" help:"hostname for docker to set"`
	Env               []string `optional:"" short:"e" help:"additional environment variable names to be passed to container"`
}

func (cmd *StartCmd) Run(c *cc.CommonCtx) error {
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

	if _, err := os.Stat(path.Join(rootPath, "cfg", "config.yaml")); errors.Is(err, os.ErrNotExist) {
		return errors.Errorf("%s does not exist, please run 'topaz configure'", path.Join(rootPath, "cfg", "config.yaml"))
	}

	generator := config.NewGenerator("config.yaml")
	if _, err := generator.CreateCertsDir(); err != nil {
		return err
	}

	if _, err := generator.CreateDataDir(); err != nil {
		return err
	}

	return dockerx.DockerWith(cmd.env(), args...)
}

var (
	dockerCmd = []string{
		"run",
	}

	dockerArgs = []string{
		"--rm",
		"--name", dockerx.Topaz,
		"--platform=linux/amd64",
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

	platform = []string{
		"--platform", "$CONTAINER_PLATFORM",
	}
)

func (cmd *StartCmd) dockerArgs(rootPath string) ([]string, error) {
	args := append([]string{}, dockerCmd...)

	policyRoot := dockerx.PolicyRoot()
	dockerArgs = append(dockerArgs, "-v", fmt.Sprintf("%s:/root/.policy:ro", policyRoot))
	args = append(args, dockerArgs...)
	args = append(args, daemonArgs...)

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
		args = append(args, hostname...)
	}

	return append(args, containerName...), nil
}

func (cmd *StartCmd) env() map[string]string {
	return map[string]string{
		"TOPAZ_CERTS_DIR":    cc.GetTopazCertsDir(),
		"TOPAZ_CFG_DIR":      cc.GetTopazCfgDir(),
		"TOPAZ_DB_DIR":       cc.GetTopazDataDir(),
		"CONTAINER_NAME":     cmd.ContainerName,
		"CONTAINER_VERSION":  cmd.ContainerVersion,
		"CONTAINER_PLATFORM": cmd.ContainerPlatform,
		"CONTAINER_HOSTNAME": cmd.ContainerHostname,
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
