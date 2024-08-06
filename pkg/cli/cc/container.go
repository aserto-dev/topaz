package cc

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
	ver "github.com/aserto-dev/topaz/pkg/version"
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
		if v.String() == "0.0.0" || v.Prerelease() != "" {
			return defaultContainerTagFallback
		}
		return v.String()
	}
	return defaultContainerTagFallback
}

// Container returns the fully qualified container name (registry/image:tag).
func Container(registry, image, tag string) string {
	if container := os.Getenv("CONTAINER"); container != "" {
		return container
	}

	return fmt.Sprintf("%s/%s:%s",
		lo.Ternary(registry != "", registry, ContainerRegistry()),
		lo.Ternary(image != "", image, ContainerImage()),
		lo.Ternary(tag != "", tag, ContainerTag()),
	)
}

// ContainerRegistry returns the container registry (host[:port]/repo).
func ContainerRegistry() string {
	if containerRegistry := os.Getenv("CONTAINER_REGISTRY"); containerRegistry != "" {
		return containerRegistry
	}
	if defaults.ContainerRegistry != "" {
		return defaults.ContainerRegistry
	}
	return defaultContainerRegistry
}

// ContainerImage returns the container image name.
func ContainerImage() string {
	if containerImage := os.Getenv("CONTAINER_IMAGE"); containerImage != "" {
		return containerImage
	}
	if defaults.ContainerImage != "" {
		return defaults.ContainerImage
	}
	return defaultContainerImage
}

// ContainerTag returns the container tag (label or semantic version).
func ContainerTag() string {
	if containerTag := os.Getenv("CONTAINER_TAG"); containerTag != "" {
		return containerTag
	}
	if defaults.ContainerTag != "" {
		return defaults.ContainerTag
	}

	return defaultContainerTag()
}

// ContainerPlatform, returns the container platform for multi-platform capable servers.
func ContainerPlatform() string {
	if containerPlatform := os.Getenv("CONTAINER_PLATFORM"); containerPlatform != "" {
		return containerPlatform
	}
	if defaults.ContainerPlatform != "" {
		return defaults.ContainerPlatform
	}
	return "linux/" + runtime.GOARCH
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
		fmt.Fprintf(os.Stderr, "!!! --container-version incl $CONTAINER_VERSION are obsolete !!!\n")
		fmt.Fprintf(os.Stderr, "instead use: --container-tag and/or $CONTAINER_TAG\n")
		fmt.Fprintf(os.Stderr, "the current --container-version value %q is propagated to --container-tag\n", version)
		return version
	}
	return tag
}
