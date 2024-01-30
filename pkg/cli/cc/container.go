package cc

import (
	"fmt"
	"os"
	"runtime"

	"github.com/aserto-dev/topaz/pkg/cli/g"
)

const (
	DefaultValue                 string = ""
	defaultContainerInstanceName string = "topaz"
)

// ContainerImage.
func ContainerImage(service, org, name, version string) string {
	if containerImage := os.Getenv("CONTAINER_IMAGE"); containerImage != "" {
		return containerImage
	}

	return fmt.Sprintf("%s/%s/%s:%s",
		g.Iff(service != "", service, ContainerRegistry()),
		g.Iff(org != "", org, ContainerOrg()),
		g.Iff(name != "", name, ContainerName()),
		g.Iff(version != "", version, ContainerVTag()),
	)
}

// ContainerRegistry.
func ContainerRegistry() string {
	if containerService := os.Getenv("CONTAINER_SERVICE"); containerService != "" {
		return containerService
	}
	return "ghcr.io"
}

// ContainerOrg.
func ContainerOrg() string {
	if containerOrg := os.Getenv("CONTAINER_ORG"); containerOrg != "" {
		return containerOrg
	}
	return "aserto-dev"
}

// ContainerName.
func ContainerName() string {
	if containerName := os.Getenv("CONTAINER_NAME"); containerName != "" {
		return containerName
	}
	return "topaz"
}

// ContainerVTag.
func ContainerVTag() string {
	if containerVersion := os.Getenv("CONTAINER_VERSION"); containerVersion != "" {
		return containerVersion
	}
	return "latest"
}

// ContainerPlatform.
func ContainerPlatform() string {
	if containerPlatform := os.Getenv("CONTAINER_PLATFORM"); containerPlatform != "" {
		return containerPlatform
	}
	return "linux/" + runtime.GOARCH
}

// ContainerInstanceName.
func ContainerInstanceName() string {
	return defaultContainerInstanceName
}
