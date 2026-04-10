package common_test

import (
	"context"
	"fmt"
	"runtime"

	"github.com/magefile/mage/sh"
	"github.com/testcontainers/testcontainers-go"
)

func GoARCH() *string {
	var goarch string
	if runtime.GOARCH == "amd64" {
		goarch = runtime.GOARCH + "_v1"
	} else {
		goarch = runtime.GOARCH + "_v8"
	}

	return &goarch
}

func CommitSHA() string {
	if commitSHA, err := sh.Output("git", "rev-parse", "--short", "HEAD"); err == nil {
		return commitSHA
	}

	return ""
}

func TestImage() string {
	return "ghcr.io/aserto-dev/topaz:0.0.0-test-" + CommitSHA() + "-" + runtime.GOARCH
}

func MappedAddr(ctx context.Context, container testcontainers.Container, port string) (string, error) {
	host, err := container.Host(ctx)
	if err != nil {
		return "", err
	}

	mappedPort, err := container.MappedPort(ctx, port)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%s", host, mappedPort.Port()), nil
}
