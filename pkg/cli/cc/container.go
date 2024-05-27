package cc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
)

const (
	defaultContainerRegistry    string = "ghcr.io/aserto-dev"
	defaultContainerImage       string = "topaz"
	defaultContainerTagFallback string = "latest"
	defaultContainerName        string = "topaz"
)

func defaultContainerTag() string {
	v, err := semver.NewVersion(ver.GetInfo().Version)
	if err == nil {
		return v.String()
	}
	return defaultContainerTagFallback
}

// Container returns the fully qualified container name (registry/image:tag).
func (c *CommonCtx) Container(registry, image, tag string) string {
	if container := os.Getenv("CONTAINER"); container != "" {
		return container
	}

	return fmt.Sprintf("%s/%s:%s",
		lo.Ternary(registry != "", registry, c.ContainerRegistry()),
		lo.Ternary(image != "", image, c.ContainerImage()),
		lo.Ternary(tag != "", tag, c.ContainerTag()),
	)
}

// ContainerRegistry returns the container registry (host[:port]/repo).
func (c *CommonCtx) ContainerRegistry() string {
	if containerRegistry := os.Getenv("CONTAINER_REGISTRY"); containerRegistry != "" {
		return containerRegistry
	}
	return c.Config.Defaults.ContainerRegistry
}

// ContainerImage returns the container image name.
func (c *CommonCtx) ContainerImage() string {
	if containerImage := os.Getenv("CONTAINER_IMAGE"); containerImage != "" {
		return containerImage
	}
	return c.Config.Defaults.ContainerImage
}

// ContainerTag returns the container tag (label or semantic version).
func (c *CommonCtx) ContainerTag() string {
	if containerTag := os.Getenv("CONTAINER_TAG"); containerTag != "" {
		return containerTag
	}

	v, err := semver.NewVersion(ver.GetInfo().Version)
	if err != nil {
		return defaultContainerTag()
	}
	return v.String()
}

// ContainerPlatform, returns the container platform for multi-platform capable servers.
func (c *CommonCtx) ContainerPlatform() string {
	if containerPlatform := os.Getenv("CONTAINER_PLATFORM"); containerPlatform != "" {
		return containerPlatform
	}
	return c.Config.Defaults.ContainerPlatform
}

// ContainerName returns the container instance name (docker run --name CONTAINER_NAME).
func ContainerName(defaultConfigFile string) string {
	if containerName := os.Getenv("CONTAINER_NAME"); containerName != "" {
		return containerName
	}
	if strings.Contains(defaultConfigFile, "config.yaml") {
		return defaultContainerName
	}
	return fmt.Sprintf("%s-%s", defaultContainerName, strings.Split(filepath.Base(defaultConfigFile), ".")[0])
}

// ContainerVersionTag consolidates the old --container-version with the --container-tag value,
// the command handlers will read the environment variable versions $CONTAINER_VERSION and $CONTAINER_TAG,
// which is why they are not explicitly handled in this function.
func ContainerVersionTag(version, tag string) string {
	if version != "" && version != tag {
		fmt.Fprintf(os.Stderr, "!!! --container-version incl $CONTAINER_VERSION are obsolete, use: --container-tag and $CONTAINER_TAG instead\n")
		fmt.Fprintf(os.Stderr, "!!! --container-version value %q it propagated to --container-tag\n", version)
		return version
	}
	return tag
}
