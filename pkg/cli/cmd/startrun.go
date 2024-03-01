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
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type StartRunCmd struct {
	ContainerRegistry string   `optional:"" default:"${container_registry}" env:"CONTAINER_REGISTRY" help:"container registry (host[:port]/repo)"`
	ContainerImage    string   `optional:"" default:"${container_image}" env:"CONTAINER_IMAGE" help:"container image name"`
	ContainerTag      string   `optional:"" default:"${container_tag}" env:"CONTAINER_TAG" help:"container tag"`
	ContainerPlatform string   `optional:"" default:"${container_platform}" env:"CONTAINER_PLATFORM" help:"container platform"`
	ContainerName     string   `optional:"" default:"${container_name}" env:"CONTAINER_NAME" help:"container name"`
	ContainerHostname string   `optional:"" name:"hostname" default:"" env:"CONTAINER_HOSTNAME" help:"hostname for docker to set"`
	Env               []string `optional:"" short:"e" help:"additional environment variable names to be passed to container"`
	ContainerVersion  string   `optional:"" hidden:"" default:"" env:"CONTAINER_VERSION"`
}

type runMode int

const (
	modeDaemon runMode = iota
	modeInteractive
)

func (cmd *StartRunCmd) run(c *cc.CommonCtx, mode runMode) error {
	if c.CheckRunStatus(cmd.ContainerName, cc.StatusRunning) {
		return ErrIsRunning
	}

	cfg, err := config.LoadConfiguration(filepath.Join(cc.GetTopazCfgDir(), "config.yaml"))
	if err != nil {
		return err
	}

	color.Green(">>> starting topaz...")
	args, err := cmd.dockerArgs(cfg, mode)
	if err != nil {
		return err
	}

	cmdArgs := []string{
		"run",
		"--config-file", "/config/config.yaml",
	}

	args = append(args, cmdArgs...)

	if _, err := os.Stat(path.Join(cc.GetTopazCfgDir(), "config.yaml")); errors.Is(err, os.ErrNotExist) {
		return errors.Errorf("%s does not exist, please run 'topaz configure'", path.Join(cc.GetTopazCfgDir(), "config.yaml"))
	}

	generator := config.NewGenerator("config.yaml")
	if _, err := generator.CreateCertsDir(); err != nil {
		return err
	}

	if _, err := generator.CreateDataDir(); err != nil {
		return err
	}

	return dockerx.DockerV(args...)
}

func (cmd *StartRunCmd) dockerArgs(cfg *config.Loader, mode runMode) ([]string, error) {
	cmd.ContainerTag = cc.ContainerVersionTag(cmd.ContainerVersion, cmd.ContainerTag)

	args := []string{
		"run",
		"--rm",
		"--name", cmd.ContainerName,
		lo.Ternary(mode == modeInteractive, "-ti", "-d"),
	}

	policyRoot := dockerx.PolicyRoot()
	args = append(args, "-v", fmt.Sprintf("%s:/root/.policy:ro", policyRoot))

	volumes, err := getVolumes(cfg)
	if err != nil {
		return nil, err
	}
	args = append(args, volumes...)

	for i := range volumes {
		if volumes[i] == "-v" {
			continue
		}
		destination := strings.Split(volumes[i], ":")
		mountedPath := fmt.Sprintf("/%s", filepath.Base(destination[1])) // last value from split.
		switch {
		case strings.Contains(volumes[i], "certs"):
			cmd.Env = append(cmd.Env, fmt.Sprintf("TOPAZ_CERTS_DIR=%s", mountedPath))
		case strings.Contains(volumes[i], "db"):
			cmd.Env = append(cmd.Env, fmt.Sprintf("TOPAZ_DB_DIR=%s", mountedPath))
		case strings.Contains(volumes[i], "cfg"):
			cmd.Env = append(cmd.Env, fmt.Sprintf("TOPAZ_CFG_DIR=%s", mountedPath))
		}
	}

	ports, err := getPorts(cfg)
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
		cc.Container(
			cmd.ContainerRegistry, // registry
			cmd.ContainerImage,    // image
			cmd.ContainerTag,      // tag
		),
	), nil
}

func getPorts(cfg *config.Loader) ([]string, error) {
	portArray, err := cfg.GetPorts()
	if err != nil {
		return nil, err
	}

	// ensure unique assignment for each port
	portMap := lo.Associate(portArray, func(port string) (string, string) {
		return port, fmt.Sprintf("%s:%s", port, port)
	})

	var args []string
	for _, v := range portMap {
		args = append(args, "-p", v)
	}
	return args, nil
}

func getVolumes(cfg *config.Loader) ([]string, error) {
	paths, err := cfg.GetPaths()
	if err != nil {
		return nil, err
	}

	volumeMap := lo.Associate(paths, func(path string) (string, string) {
		dir := filepath.Dir(path)
		return dir, fmt.Sprintf("%s:%s", dir, fmt.Sprintf("/%s", filepath.Base(dir)))
	})

	// manually attach the configuration folder
	args := []string{"-v", fmt.Sprintf("%s:/config:ro", cc.GetTopazCfgDir())}
	for _, v := range volumeMap {
		args = append(args, "-v", v)
	}
	return args, nil
}
