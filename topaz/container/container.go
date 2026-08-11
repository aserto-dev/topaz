package container

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aserto-dev/topaz/topaz/x"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func PolicyRoot() string {
	const defaultPolicyRoot = ".policy"

	policyRoot := os.Getenv(x.EnvPolicyFileStoreRoot)
	if policyRoot == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}

		policyRoot = filepath.Join(home, defaultPolicyRoot)
	}

	//nolint:gosec // G703
	if fi, err := os.Stat(policyRoot); err == nil && fi.IsDir() {
		return policyRoot
	}

	return ""
}

type ContainerClient interface {
	PullImage(img, platform string) error
	RemoveImage(img string) error
	ImageExists(img string) bool
	Stop(name string) error
	IsRunning(name string) (bool, error)
	GetRunningTopazContainers() ([]container.Summary, error)
	Run(opts ...RunOption) error
	Start(opts ...RunOption) error
}

type ContainerEngine int64

const (
	Docker    ContainerEngine = iota // Docker container, supports AMD64 & ARM64.
	Container                        // Apple Container, supports ARM64 only.
)

func New(engine ContainerEngine) (*ContainerClient, error) {
	return nil, nil
}

type runner struct {
	config           *container.Config
	hostConfig       *container.HostConfig
	networkingConfig *network.NetworkingConfig
	platform         *v1.Platform
	containerName    string
	runOut           io.Writer
	runErr           io.Writer
}

type RunOption func(*runner)

func WithContainerImage(img string) RunOption {
	return func(r *runner) {
		r.config.Image = img
	}
}

func WithWorkingDir(wd string) RunOption {
	return func(r *runner) {
		r.config.WorkingDir = wd
	}
}

func WithEntrypoint(args []string) RunOption {
	return func(r *runner) {
		r.config.Entrypoint = args
	}
}

func WithCmd(cmds []string) RunOption {
	return func(r *runner) {
		r.config.Cmd = cmds
	}
}

func WithContainerPlatform(platform string) RunOption {
	goos, goarch := "linux", strings.TrimPrefix(platform, "linux/")

	return func(r *runner) {
		r.platform.OS = goos
		r.platform.Architecture = goarch
	}
}

func WithContainerName(name string) RunOption {
	return func(r *runner) {
		r.containerName = name
	}
}

func WithContainerHostname(hostname string) RunOption {
	return func(r *runner) {
		r.config.Hostname = hostname
	}
}

// WithAutoRemove, automatically remove container when it exits.
func WithAutoRemove() RunOption {
	return func(r *runner) {
		r.hostConfig.AutoRemove = true
	}
}

func WithEnv(env string) RunOption {
	return func(r *runner) {
		r.config.Env = append(r.config.Env, env)
	}
}

func WithEnvs(envs []string) RunOption {
	return func(r *runner) {
		r.config.Env = append(r.config.Env, envs...)
	}
}

func WithPort(port string) RunOption {
	return func(r *runner) {
		_ = r.setPorts([]string{port})
	}
}

func WithPorts(ports []string) RunOption {
	return func(r *runner) {
		_ = r.setPorts(ports)
	}
}

func (r *runner) setPorts(ports []string) error {
	portSet, portBindings, err := nat.ParsePortSpecs(ports)
	if err != nil {
		return err
	}

	if r.config.ExposedPorts == nil {
		r.config.ExposedPorts = make(nat.PortSet)
	}

	for port, v := range portSet {
		if _, ok := r.config.ExposedPorts[port]; !ok {
			r.config.ExposedPorts[port] = v
		}
	}

	if r.hostConfig.PortBindings == nil {
		r.hostConfig.PortBindings = make(nat.PortMap)
	}

	for port, binding := range portBindings {
		if _, ok := r.hostConfig.PortBindings[port]; !ok {
			r.hostConfig.PortBindings[port] = binding
		}
	}

	return nil
}

func WithVolume(volume string) RunOption {
	return func(r *runner) {
		r.hostConfig.Binds = append(r.hostConfig.Binds, volume)
	}
}

func WithVolumes(volumes []string) RunOption {
	return func(r *runner) {
		r.hostConfig.Binds = append(r.hostConfig.Binds, volumes...)
	}
}

func WithOutput(o io.Writer) RunOption {
	return func(r *runner) {
		r.runOut = o
	}
}

func WithError(e io.Writer) RunOption {
	return func(r *runner) {
		r.runErr = e
	}
}
