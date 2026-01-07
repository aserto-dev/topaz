package topaz

import (
	"fmt"
	"iter"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	cfgutil "github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/pkg/config/v3"
	"github.com/aserto-dev/topaz/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/topaz/pkg/cli/cmd/common"
	clicfg "github.com/aserto-dev/topaz/topaz/pkg/cli/config"
	"github.com/aserto-dev/topaz/topaz/pkg/cli/dockerx"
	"github.com/aserto-dev/topaz/topaz/pkg/cli/x"
	"github.com/aserto-dev/topaz/topazd/loiter"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

var ErrUndefinedEnvVar = errors.New("environment variable is referenced but not defined")

type StartRunCmd struct {
	ContainerRegistry string   `optional:"" default:"${container_registry}" env:"CONTAINER_REGISTRY" help:"container registry (host[:port]/repo)"`
	ContainerImage    string   `optional:"" default:"${container_image}" env:"CONTAINER_IMAGE" help:"container image name"`
	ContainerTag      string   `optional:"" default:"${container_tag}" env:"CONTAINER_TAG" help:"container tag"`
	ContainerPlatform string   `optional:"" default:"${container_platform}" env:"CONTAINER_PLATFORM" help:"container platform"`
	ContainerName     string   `optional:"" default:"${container_name}" env:"CONTAINER_NAME" help:"container name"`
	ContainerHostname string   `optional:"" name:"hostname" default:"" env:"CONTAINER_HOSTNAME" help:"hostname for docker to set"`
	Env               []string `optional:"" short:"e" help:"additional environment variable names to be passed to container"`
	ContainerVersion  string   `optional:"" hidden:"" default:"" env:"CONTAINER_VERSION"`

	cfg        *clicfg.Container `kong:"-"`
	envMapping map[string]string `kong:"-"`
}

type runMode int

const (
	modeDaemon runMode = iota
	modeInteractive
)

func containerPorts(ports iter.Seq[string]) []string {
	return slices.Collect(
		loiter.Map(ports, func(port string) string {
			return fmt.Sprintf("%s:%s/tcp", port, port)
		}),
	)
}

type mount struct {
	src  string
	dest string
	mode cfgutil.AccessMode
}

func (m mount) String() string {
	return fmt.Sprintf("%s:%s:%s", m.src, m.dest, m.mode)
}

func (cmd *StartRunCmd) ExpandHostPath(path string) string {
	return os.Expand(path, func(name string) string {
		switch name {
		case x.EnvTopazCertsDir:
			return cc.GetTopazCertsDir()
		case x.EnvTopazCfgDir:
			return cc.GetTopazCfgDir()
		case x.EnvTopazDBDir:
			return cc.GetTopazDataDir()
		default:
			return os.Getenv(name)
		}
	})
}

func (cmd *StartRunCmd) ExpandContainerPath(path string) string {
	return os.Expand(path, func(name string) string {
		if value, ok := cmd.envMapping[name]; ok {
			return value
		}

		switch name {
		case x.EnvTopazCertsDir:
			return x.DefCertsDir
		case x.EnvTopazCfgDir:
			return x.DefCfgDir
		case x.EnvTopazDBDir:
			return x.DefDBDir
		default:
			return ""
		}
	})
}

func (cmd *StartRunCmd) run(c *cc.CommonCtx, mode runMode) error {
	if c.CheckRunStatus(cmd.ContainerName, cc.StatusRunning) {
		return cc.ErrIsRunning
	}

	if _, err := os.Stat(c.Config.Active.ConfigFile); errors.Is(err, os.ErrNotExist) {
		return errors.Errorf("%s does not exist, please run 'topaz config new'", path.Join(c.Config.Active.ConfigFile))
	}

	cfg, err := cmd.loadConfig(c.Config.Active.ConfigFile)
	if err != nil {
		return err
	}

	if err := cmd.resolveEnv(); err != nil {
		// Should this be a warning instead of an error?
		return err
	}

	c.Config.Running.Config = c.Config.Active.Config
	c.Config.Running.ConfigFile = c.Config.Active.ConfigFile
	c.Config.Running.ContainerName = cc.ContainerName(c.Config.Active.ConfigFile)

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

	volumes := cmd.volumeMounts(cfg.Config, c.Config.Active.ConfigFile)
	env := cmd.containerEnv()

	opts := []dockerx.RunOption{
		dockerx.WithContainerImage(image),
		dockerx.WithContainerPlatform(cmd.ContainerPlatform),
		dockerx.WithContainerName(cmd.ContainerName),
		dockerx.WithContainerHostname(cmd.ContainerHostname),
		dockerx.WithWorkingDir("/app"),
		dockerx.WithEntrypoint([]string{"./topazd"}),
		dockerx.WithCmd([]string{"run", "--config-file", "/config/" + filepath.Base(c.Config.Active.ConfigFile)}),
		dockerx.WithAutoRemove(),
		dockerx.WithEnvs(env),
		dockerx.WithPorts(containerPorts(cfg.Ports())),
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

func (cmd *StartRunCmd) loadConfig(path string) (*clicfg.Container, error) {
	cfg, err := clicfg.Load(path)

	cmd.cfg = cfg

	return cfg, err
}

// resolveEnv populates cmd.envMapping with the environment variables defined in cmd.Env.
// Variables that don't provide a value are looked up in the environment and an error is returned if they are not
// defined.
func (cmd *StartRunCmd) resolveEnv() error {
	var errs error

	cmd.envMapping = lo.Associate(cmd.Env, func(env string) (string, string) {
		name, value, _ := strings.Cut(env, "=")
		if value == "" {
			// Value is not provided. Look it up in the environment.
			if v, ok := os.LookupEnv(name); ok {
				value = v
			} else {
				errs = multierror.Append(errs, errors.Wrap(ErrUndefinedEnvVar, name))
			}
		}

		return name, value
	})

	if _, ok := cmd.envMapping[x.EnvTopazCertsDir]; !ok {
		cmd.envMapping[x.EnvTopazCertsDir] = x.DefCertsDir
	}

	if _, ok := cmd.envMapping[x.EnvTopazCfgDir]; !ok {
		cmd.envMapping[x.EnvTopazCfgDir] = x.DefCfgDir
	}

	if _, ok := cmd.envMapping[x.EnvTopazDBDir]; !ok {
		cmd.envMapping[x.EnvTopazDBDir] = x.DefDBDir
	}

	return errs
}

func (cmd *StartRunCmd) containerEnv() []string {
	return lo.MapToSlice(cmd.envMapping, func(name, value string) string {
		if value == "" {
			return name
		}

		return fmt.Sprintf("%s=%s", name, value)
	})
}

func (cmd *StartRunCmd) volumeMounts(cfg *config.Config, cfgFile string) []string {
	paths := loiter.Collect(
		loiter.RejectKey(cfg.Paths(), ""),
		func(_ string, _, mode cfgutil.AccessMode) bool { return mode == cfgutil.ReadWrite },
	)

	return append(
		lo.MapToSlice(paths, func(path string, mode cfgutil.AccessMode) string {
			return mount{
				src:  cmd.ExpandHostPath(path),
				dest: cmd.ExpandContainerPath(path),
				mode: mode,
			}.String()
		}),
		fmt.Sprintf("%s:%s:ro", cfgFile, path.Join(x.DefCfgDir, filepath.Base(cfgFile))), // always mount the config file as read-only.
	)
}
