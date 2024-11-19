package common_test

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/magefile/mage/sh"
	"github.com/testcontainers/testcontainers-go"
)

type Harness struct {
	container testcontainers.Container
}

func NewHarness(ctx context.Context, req *testcontainers.ContainerRequest) (*Harness, error) {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: *req,
		Started:          false,
	})
	if err != nil {
		return nil, err
	}

	if err := container.Start(ctx); err != nil {
		return nil, err
	}

	return &Harness{
		container: container,
	}, nil
}

func (h *Harness) Close(ctx context.Context) error {
	timeout := 20 * time.Second
	if err := h.container.Stop(ctx, &timeout); err != nil {
		return h.container.Terminate(ctx)
	}
	return nil
}

func (h *Harness) AddrGRPC(ctx context.Context) string {
	host, err := h.container.Host(ctx)
	if err != nil {
		log.Fatal(err)
	}

	mappedPort, err := h.container.MappedPort(ctx, "9292")
	if err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%s:%s", host, mappedPort.Port())
}

func (h *Harness) AddrREST(ctx context.Context) string {
	host, err := h.container.Host(ctx)
	if err != nil {
		log.Fatal(err)
		// return ""
	}

	mappedPort, err := h.container.MappedPort(ctx, "9393")
	if err != nil {
		log.Fatal(err)
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
	return "ghcr.io/aserto-dev/topaz:0.0.0-test-" + CommitSHA() + "-" + runtime.GOARCH
}
