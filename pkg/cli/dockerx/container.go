package dockerx

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var ErrNotRunning = errors.New("topaz is not running, use 'topaz start' or 'topaz run' to start")

type RunMode bool

const (
	Interactive RunMode = true
	Deamon      RunMode = false
)

type Container struct {
	ContainerName    string   `optional:"" default:"topaz" help:"container name"`
	ContainerVersion string   `optional:"" default:"latest" help:"container version" `
	Hostname         string   `optional:"" help:"hostname for docker to set"`
	Env              []string `optional:"" short:"e" help:"additional environment variable names to be passed to container"`
}

func (c *Container) env(rootPath string) map[string]string {
	return map[string]string{
		"TOPAZ_CERTS_DIR":    rootPath,
		"TOPAZ_CFG_DIR":      rootPath,
		"TOPAZ_EDS_DIR":      rootPath,
		"CONTAINER_NAME":     c.ContainerName,
		"CONTAINER_VERSION":  c.ContainerVersion,
		"CONTAINER_HOSTNAME": c.Hostname,
	}
}

func (c *Container) dockerArgs(localBundles []string, mode RunMode) []string {
	args := append([]string{}, dockerCmd...)
	args = append(args, dockerArgs...)
	switch mode {
	case Interactive:
		args = append(args, "-ti")
	case Deamon:
		args = append(args, "-d")
	}

	args = append(args, localBundleMounts(localBundles)...)

	for _, env := range c.Env {
		args = append(args, "--env", env)
	}

	if c.Hostname != "" {
		args = append(args, Hostname...)
	}

	return append(args, ImageName...)
}

func (c *Container) Start(mode RunMode) error {
	if running, err := IsRunning(Topaz); running || err != nil {
		if !running {
			return ErrNotRunning
		}
		if err != nil {
			return err
		}
	}

	rootPath, err := DefaultRoots()
	if err != nil {
		return err
	}

	configFile := path.Join(rootPath, "cfg", "config.yaml")
	localBundles, err := parseLocalBundles(configFile)
	if err != nil {
		return err
	}

	color.Green(">>> starting topaz...")

	args := c.dockerArgs(localBundles, mode)

	cmdArgs := []string{
		"run",
		"--config-file", "/config/config.yaml",
	}

	args = append(args, cmdArgs...)

	return DockerWith(c.env(rootPath), args...)

}

var (
	dockerCmd = []string{
		"run",
	}

	dockerArgs = []string{
		"--rm",
		"--name", Topaz,
		"--platform=linux/amd64",
		"-p", "8282:8282",
		"-p", "8383:8383",
		"-p", "8484:8484",
		"-p", "9292:9292",
		"-v", "$TOPAZ_CERTS_DIR/certs:/certs:rw",
		"-v", "$TOPAZ_CFG_DIR/cfg/config.yaml:/config/config.yaml:ro",
		"-v", "$TOPAZ_EDS_DIR/db:/db:rw",
	}

	ImageName = []string{
		"ghcr.io/aserto-dev/$CONTAINER_NAME:$CONTAINER_VERSION",
	}

	Hostname = []string{
		"--hostname", "$CONTAINER_HOSTNAME",
	}

	Platform = []string{
		"--platform", "linux/amd64",
	}
)

type topazConfig struct {
	OPA struct {
		LocalBundles struct {
			Paths []string `yaml:"paths"`
		} `yaml:"local_bundles"`
	} `yaml:"opa"`
}

func parseLocalBundles(configFile string) ([]string, error) {
	bundles := []string{}

	f, err := os.Open(configFile)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return bundles, errors.Errorf("%s does not exist, please run 'topaz configure'", configFile)
	case err != nil:
		return bundles, errors.Wrapf(err, "failed to open %s", configFile)
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return bundles, errors.Wrapf(err, "failed to read %s", configFile)
	}

	var cfg topazConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return bundles, errors.Wrapf(err, "failed to parse %s", configFile)
	}

	return cfg.OPA.LocalBundles.Paths, nil
}

func localBundleMounts(localBundles []string) []string {
	mounts := []string{}
	for _, bundle := range localBundles {
		mounts = append(mounts, "-v", fmt.Sprintf("%s:%s:ro", bundle, bundle))
	}
	return mounts
}
