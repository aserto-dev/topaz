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
	ContainerName    string   `optional:"" default:"topaz" help:"container name"`
	ContainerVersion string   `optional:"" default:"latest" help:"container version" `
	Hostname         string   `optional:"" help:"hostname for docker to set"`
	Env              []string `optional:"" short:"e" help:"additional environment variable names to be passed to container"`
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

	return dockerx.DockerWith(cmd.env(rootPath), args...)
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
		"--platform", "linux/amd64",
	}
)

func (cmd *StartCmd) dockerArgs(rootPath string) ([]string, error) {
	args := append([]string{}, dockerCmd...)

	policyRoot := dockerx.PolicyRoot()
	dockerArgs = append(dockerArgs, "-v", fmt.Sprintf("%s:/root/.policy:ro", policyRoot))
	args = append(args, dockerArgs...)
	args = append(args, daemonArgs...)

	for _, env := range cmd.Env {
		args = append(args, "--env", env)
	}

	volumes, err := getVolumes(rootPath)
	if err != nil {
		return nil, err
	}
	args = append(args, volumes...)

	ports, err := getPorts(rootPath)
	if err != nil {
		return nil, err
	}
	args = append(args, ports...)

	if cmd.Hostname != "" {
		args = append(args, hostname...)
	}

	return append(args, containerName...), nil
}

func (cmd *StartCmd) env(rootPath string) map[string]string {
	return map[string]string{
		"TOPAZ_CERTS_DIR":    rootPath,
		"TOPAZ_CFG_DIR":      rootPath,
		"TOPAZ_EDS_DIR":      rootPath,
		"CONTAINER_NAME":     cmd.ContainerName,
		"CONTAINER_VERSION":  cmd.ContainerVersion,
		"CONTAINER_HOSTNAME": cmd.Hostname,
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
	configLoader, err := config.LoadConfiguration(fmt.Sprintf("%s/cfg/config.yaml", rootPath))
	if err != nil {
		return nil, err
	}

	paths, err := configLoader.GetPaths()
	if err != nil {
		return nil, err
	}

	for i := range paths {

		mountPath := strings.Split(paths[i], string(os.PathSeparator))
		mountPath[0] = "" // replace root path from generated config for container
		directory := filepath.Dir(paths[i])
		volumeMap[directory] = fmt.Sprintf("%s:%s", directory, filepath.Dir(strings.Join(mountPath, "/")))
	}

	// manually attach the configuration folder
	args := []string{"-v", "$TOPAZ_CFG_DIR/cfg:/config:ro"}
	for _, v := range volumeMap {
		args = append(args, "-v", v)
	}
	return args, nil
}
