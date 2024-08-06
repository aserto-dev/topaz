package topaz

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/common"
	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
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
		return cc.ErrIsRunning
	}

	if _, err := os.Stat(c.Config.Active.ConfigFile); errors.Is(err, os.ErrNotExist) {
		return errors.Errorf("%s does not exist, please run 'topaz config new'", path.Join(c.Config.Active.ConfigFile))
	}

	cfg, err := config.LoadConfiguration(c.Config.Active.ConfigFile)
	if err != nil {
		return err
	}

	c.Config.Running.Config = c.Config.Active.Config
	c.Config.Running.ConfigFile = c.Config.Active.ConfigFile
	c.Config.Running.ContainerName = cc.ContainerName(c.Config.Active.ConfigFile)

	if cfg.HasTopazDir {
		c.Con().Warn().Msg("This configuration file still uses TOPAZ_DIR environment variable.")
		c.Con().Msg("Please change to using the new TOPAZ_DB_DIR and TOPAZ_CERTS_DIR environment variables.")
	}

	generator := config.NewGenerator(filepath.Base(c.Config.Active.ConfigFile))
	if _, err := generator.CreateCertsDir(); err != nil {
		return err
	}

	if _, err := generator.CreateDataDir(); err != nil {
		return err
	}

	ports, err := getPorts(cfg)
	if err != nil {
		return err
	}

	volumes, err := getVolumes(cfg)
	if err != nil {
		return err
	}

	c.Con().Info().Msg(">>> starting topaz %q...", c.Config.Running.Config)

	dc, err := dockerx.New()
	if err != nil {
		return err
	}

	image := cc.Container(
		cmd.ContainerRegistry, // registry
		cmd.ContainerImage,    // image
		cmd.ContainerTag,      // tag
	)

	if !dc.ImageExists(image) {
		if err := dc.PullImage(image, cmd.ContainerPlatform); err != nil {
			return err
		}
	}

	opts := []dockerx.RunOption{
		dockerx.WithContainerImage(image),
		dockerx.WithContainerPlatform(cmd.ContainerPlatform),
		dockerx.WithContainerName(cmd.ContainerName),
		dockerx.WithContainerHostname(cmd.ContainerHostname),
		dockerx.WithWorkingDir("/app"),
		dockerx.WithEntrypoint([]string{"./topazd"}),
		dockerx.WithCmd([]string{"run", "--config-file", fmt.Sprintf("/config/%s", filepath.Base(c.Config.Active.ConfigFile))}),
		dockerx.WithAutoRemove(),
		dockerx.WithEnvs(getEnvFromVolumes(volumes)),
		dockerx.WithEnvs(cmd.Env),
		dockerx.WithPorts(ports),
		dockerx.WithVolumes(volumes),
		dockerx.WithOutput(c.StdOut()),
		dockerx.WithError(c.StdErr()),
	}

	if mode == modeInteractive {
		if err := dc.Run(opts...); err != nil {
			return err
		}
	}

	if mode == modeDaemon {
		if err := dc.Start(opts...); err != nil {
			return err
		}
	}

	c.Con().Msg("")

	if err := c.SaveContextConfig(common.CLIConfigurationFile); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	return nil
}

func getPorts(cfg *config.Loader) ([]string, error) {
	portArray, err := cfg.GetPorts()
	if err != nil {
		return nil, err
	}

	// ensure unique assignment for each port
	portMap := lo.Associate(portArray, func(port string) (string, string) {
		return port, fmt.Sprintf("%s:%s/tcp", port, port)
	})

	var ports []string
	for _, v := range portMap {
		ports = append(ports, v)
	}

	return ports, nil
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

	volumes := []string{
		fmt.Sprintf("%s:/config:ro", cc.GetTopazCfgDir()), // manually attach the configuration folder
	}

	if cfg.Configuration.OPA.LocalBundles.LocalPolicyImage != "" && dockerx.PolicyRoot() != "" {
		volumes = append(volumes, fmt.Sprintf("%s:/root/.policy:ro", dockerx.PolicyRoot())) // manually attach policy store
	}

	for _, v := range volumeMap {
		volumes = append(volumes, v)
	}
	return volumes, nil
}

func getEnvFromVolumes(volumes []string) []string {
	envs := []string{}
	for i := range volumes {
		destination := strings.Split(volumes[i], ":")
		mountedPath := fmt.Sprintf("/%s", filepath.Base(destination[1])) // last value from split.
		switch {
		case strings.Contains(volumes[i], "certs"):
			envs = append(envs, fmt.Sprintf("TOPAZ_CERTS_DIR=%s", mountedPath))
		case strings.Contains(volumes[i], "db"):
			envs = append(envs, fmt.Sprintf("TOPAZ_DB_DIR=%s", mountedPath))
		case strings.Contains(volumes[i], "cfg"):
			envs = append(envs, fmt.Sprintf("TOPAZ_CFG_DIR=%s", mountedPath))
		}
	}
	return envs
}
