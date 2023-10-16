package cmd

import (
	"fmt"
	"os"
	"path"

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

	color.Green(">>> starting topaz...")

	args := cmd.dockerArgs()

	cmdArgs := []string{
		"run",
		"--config-file", "/config/config.yaml",
	}

	args = append(args, cmdArgs...)

	rootPath, err := dockerx.DefaultRoots()
	if err != nil {
		return err
	}

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
		"-p", "8282:8282",
		"-p", "8383:8383",
		"-p", "8484:8484",
		"-p", "9292:9292",
		"-p", "9393:9393",
		"-p", "8080:8080",
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

func (cmd *StartCmd) dockerArgs() []string {
	args := append([]string{}, dockerCmd...)

	policyRoot := os.Getenv("POLICY_FILE_STORE_ROOT ")
	if policyRoot == "" {
		policyRoot = dockerx.DefaultPolicyRoot
	}
	dockerArgs = append(dockerArgs, "-v", fmt.Sprintf("%s:/root/.policy:ro", policyRoot))
	args = append(args, dockerArgs...)
	args = append(args, daemonArgs...)

	for _, env := range cmd.Env {
		args = append(args, "--env", env)
	}

	if cmd.Hostname != "" {
		args = append(args, hostname...)
	}

	return append(args, containerName...)
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
