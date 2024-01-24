package cc

import (
	"os"
	"runtime"
)

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
