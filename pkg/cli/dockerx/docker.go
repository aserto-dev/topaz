package dockerx

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func PolicyRoot() string {
	const defaultPolicyRoot = ".policy"

	policyRoot := os.Getenv("POLICY_FILE_STORE_ROOT")
	if policyRoot == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}

		return path.Join(home, defaultPolicyRoot)
	}
	return policyRoot
}

type DockerClient struct {
	ctx context.Context
	cli *client.Client
}

func New() (*DockerClient, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &DockerClient{
		ctx: context.Background(),
		cli: cli,
	}, nil
}

// PullImage container image.
func (dc *DockerClient) PullImage(image, platform string) error {
	out, err := dc.cli.ImagePull(dc.ctx, image, types.ImagePullOptions{
		Platform: platform,
	})
	if err != nil {
		return err
	}
	defer out.Close()

	_, _ = io.Copy(io.Discard, out)

	return nil
}

// Remove container image as image-name:tag.
func (dc *DockerClient) RemoveImage(image string) error {
	images, err := dc.cli.ImageList(dc.ctx, types.ImageListOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{
				Key: "reference", Value: image,
			},
		),
	})
	if err != nil {
		return err
	}

	for i := 0; i < len(images); i++ {
		_, err := dc.cli.ImageRemove(dc.ctx, images[i].ID, types.ImageRemoveOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

// Check if image exists in local container store.
func (dc *DockerClient) ImageExists(image string) bool {
	images, err := dc.cli.ImageList(dc.ctx, types.ImageListOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{
				Key: "reference", Value: image,
			},
		),
	})
	if err != nil {
		return false
	}

	return len(images) == 1
}

// Stop container instance with `name`.
func (dc *DockerClient) Stop(name string) error {
	containers, err := dc.cli.ContainerList(dc.ctx, container.ListOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{
				Key: "status", Value: "running"},
			filters.KeyValuePair{
				Key: "name", Value: name,
			}),
	})
	if err != nil {
		return err
	}

	waitTimeout := 10
	for i := 0; i < len(containers); i++ {
		if err := dc.cli.ContainerStop(dc.ctx, containers[i].ID, container.StopOptions{Timeout: &waitTimeout}); err != nil {
			return err
		}
	}

	return nil
}

// IsRunning returns if container with `name` is running.
func (dc *DockerClient) IsRunning(name string) (bool, error) {
	containers, err := dc.cli.ContainerList(dc.ctx, container.ListOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{
				Key: "status", Value: "running"},
			filters.KeyValuePair{
				Key: "name", Value: name,
			}),
	})
	if err != nil {
		return false, err
	}

	rc := false
	if len(containers) == 1 {
		rc = containers[0].State == "running"
	}

	return rc, nil
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

func WithContainerImage(image string) RunOption {
	return func(r *runner) {
		r.config.Image = image
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

// Run starts a container like `docker run` using the provided settings.
func (dc *DockerClient) Run(opts ...RunOption) error {
	r := &runner{
		config: &container.Config{
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			OpenStdin:    true,
			StdinOnce:    true,
			Tty:          true,
		},
		hostConfig:       &container.HostConfig{},
		networkingConfig: &network.NetworkingConfig{},
		platform:         &v1.Platform{},
	}

	for _, opt := range opts {
		opt(r)
	}

	cont, err := dc.cli.ContainerCreate(
		dc.ctx,
		r.config,
		r.hostConfig,
		r.networkingConfig,
		r.platform,
		r.containerName,
	)
	if err != nil {
		return err
	}

	if err := dc.cli.ContainerStart(dc.ctx, cont.ID, container.StartOptions{}); err != nil {
		return err
	}
	defer func() {
		_ = dc.cli.ContainerRemove(dc.ctx, cont.ID, container.RemoveOptions{Force: true})
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		_ = dc.cli.ContainerStop(dc.ctx, cont.ID, container.StopOptions{})
	}()

	go func() {
		reader, err := dc.cli.ContainerLogs(dc.ctx, cont.ID,
			container.LogsOptions{
				ShowStdout: true,
				ShowStderr: true,
				Follow:     true,
				Timestamps: false,
			})
		if err != nil {
			panic(err)
		}
		defer reader.Close()

		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	statusCh, errCh := dc.cli.ContainerWait(dc.ctx, cont.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-statusCh:
	}

	return nil
}

// Starts starts a container like `docker start` using the provided settings.
func (dc *DockerClient) Start(opts ...RunOption) error {
	r := &runner{
		config:           &container.Config{},
		hostConfig:       &container.HostConfig{},
		networkingConfig: &network.NetworkingConfig{},
		platform:         &v1.Platform{},
	}

	for _, opt := range opts {
		opt(r)
	}

	cont, err := dc.cli.ContainerCreate(
		dc.ctx,
		r.config,
		r.hostConfig,
		r.networkingConfig,
		r.platform,
		r.containerName,
	)
	if err != nil {
		return err
	}

	if err := dc.cli.ContainerStart(dc.ctx, cont.ID, container.StartOptions{}); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "%s\n", cont.ID)

	return nil
}
