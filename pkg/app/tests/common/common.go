package common_test

import (
	"context"
	"fmt"
	"runtime"

	"github.com/magefile/mage/sh"
	"github.com/testcontainers/testcontainers-go"
)

type Harness struct {
	container testcontainers.Container
}

func NewHarness(ctx context.Context, req *testcontainers.ContainerRequest) (*Harness, error) {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: *req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &Harness{
		container: container,
	}, nil
}

func (h *Harness) Close(ctx context.Context) error {
	return h.container.Terminate(ctx)
}

func (h *Harness) AddrGRPC(ctx context.Context) string {
	host, err := h.container.Host(ctx)
	if err != nil {
		return ""
	}

	mappedPort, err := h.container.MappedPort(ctx, "9292")
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s:%s", host, mappedPort.Port())
}

func (h *Harness) AddrREST(ctx context.Context) string {
	host, err := h.container.Host(ctx)
	if err != nil {
		return ""
	}

	mappedPort, err := h.container.MappedPort(ctx, "9393")
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s:%s", host, mappedPort.Port())
}

func GoARCH() *string {
	var goarch string
	if runtime.GOARCH == "amd64" {
		goarch = runtime.GOARCH + "_v1"
	} else {
		goarch = runtime.GOARCH
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
	return "ghcr.io/aserto-dev/topaz:test-" + CommitSHA() + "-" + runtime.GOARCH
}
