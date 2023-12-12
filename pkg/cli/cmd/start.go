package cmd

import (
	"fmt"
	"net"
	"os"
	"path"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

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

	if _, err := CreateCertsDir(); err != nil {
		return err
	}

	if _, err := CreateDataDir(); err != nil {
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
	content, err := os.ReadFile(fmt.Sprintf("%s/cfg/config.yaml", rootPath))
	if err != nil {
		return nil, err
	}

	var api map[string]interface{}
	err = yaml.Unmarshal(content, &api)
	if err != nil {
		return nil, err
	}

	healthConfig := getValueByKey(api["api"], "health")
	if healthConfig != nil {
		port, err := getPort(healthConfig)
		if err != nil {
			return nil, err
		}
		portMap[port] = fmt.Sprintf("%s:%s", port, port)
	}

	metricsConfig := getValueByKey(api["api"], "metrics")
	if metricsConfig != nil {
		port, err := getPort(metricsConfig)
		if err != nil {
			return nil, err
		}
		portMap[port] = fmt.Sprintf("%s:%s", port, port)
	}

	servicesConfig, ok := getValueByKey(api["api"], "services").(map[interface{}]interface{})
	if ok {
		for _, value := range servicesConfig {
			grpcConfig := getValueByKey(value, "grpc")
			if grpcConfig != nil {
				port, err := getPort(grpcConfig)
				if err != nil {
					return nil, err
				}
				portMap[port] = fmt.Sprintf("%s:%s", port, port)
			}
			gatewayConfig := getValueByKey(value, "gateway")
			if gatewayConfig != nil {
				port, err := getPort(gatewayConfig)
				if err != nil {
					return nil, err
				}
				portMap[port] = fmt.Sprintf("%s:%s", port, port)
			}
		}
	}

	// ensure unique assignment for each port
	var args []string
	for _, v := range portMap {
		args = append(args, "-p", v)
	}
	return args, nil
}

func getPort(obj interface{}) (string, error) {
	address := getValueByKey(obj, "listen_address")
	if address != nil {
		_, port, err := net.SplitHostPort(address.(string))
		if err != nil {
			return "", err
		}
		return port, nil
	}

	return "", fmt.Errorf("listen_address not found")
}

func getValueByKey(obj interface{}, key string) interface{} {
	mobj, ok := obj.(map[interface{}]interface{})
	if !ok {
		return nil
	}
	for k, v := range mobj {
		if k == key {
			return v
		}
	}
	return nil
}
