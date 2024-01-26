package cc

import (
	"fmt"
	"os"
	"runtime"

	"github.com/aserto-dev/topaz/pkg/cli/g"
)

const DefaultValue string = ""

func GetContainerImage(service, org, name, version string) string {
	if containerImage := os.Getenv("CONTAINER_IMAGE"); containerImage != "" {
		return containerImage
	}

	return fmt.Sprintf("%s/%s/%s:%s",
		g.Iff(service != "", service, GetContainerService()),
		g.Iff(org != "", org, GetContainerOrg()),
		g.Iff(name != "", name, GetContainerName()),
		g.Iff(version != "", version, GetContainerVersion()),
	)
}

func GetContainerService() string {
	if containerService := os.Getenv("CONTAINER_SERVICE"); containerService != "" {
		return containerService
	}
	return "ghcr.io"
}

func GetContainerOrg() string {
	if containerOrg := os.Getenv("CONTAINER_ORG"); containerOrg != "" {
		return containerOrg
	}
	return "aserto-dev"
}

func GetContainerName() string {
	if containerName := os.Getenv("CONTAINER_NAME"); containerName != "" {
		return containerName
	}
	return "topaz"
}

func GetContainerVersion() string {
	if containerVersion := os.Getenv("CONTAINER_VERSION"); containerVersion != "" {
		return containerVersion
	}
	return "latest"
}

func GetContainerPlatform() string {
	if containerPlatform := os.Getenv("CONTAINER_PLATFORM"); containerPlatform != "" {
		return containerPlatform
	}
	return "linux/" + runtime.GOARCH
}
