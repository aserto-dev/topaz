package dockerx

import (
	"context"
	"os"
	"path"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/magefile/mage/sh"
)

const (
	docker string = "docker"
)

var (
	DockerRun = sh.RunCmd(docker)
	DockerOut = sh.OutCmd(docker)
)

func DockerWith(env map[string]string, args ...string) error {
	return sh.RunWithV(env, docker, args...)
}

func DockerWithOut(env map[string]string, args ...string) (string, error) {
	return sh.OutputWith(env, docker, args...)
}

func IsRunning(name string) (bool, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation())
	if err != nil {
		return false, err
	}

	containers, err := cli.ContainerList(context.Background(), container.ListOptions{
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
