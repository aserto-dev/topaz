package dockerx_test

import (
	"reflect"
	"testing"

	"github.com/aserto-dev/topaz/pkg/cli/dockerx"
	"github.com/stretchr/testify/require"
)

const (
	containerName    = "container name"
	containerVersion = "version"
	hostname         = "hostname"
)

func newContainer(env ...string) *dockerx.Container {
	return &dockerx.Container{
		ContainerName:    containerName,
		ContainerVersion: containerVersion,
		Hostname:         hostname,
		Env:              env,
	}
}

func TestDockerArgs(t *testing.T) {
	tests := []struct {
		name      string
		container *dockerx.Container
		mode      dockerx.RunMode
	}{
		{"interactive", newContainer(), dockerx.Interactive},
		{"daemon", newContainer(), dockerx.Deamon},
		{"env", newContainer("env1", "env2"), dockerx.Interactive},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			args := test.container.DockerArgs(test.mode)
			validateArgs(t, test.container, test.mode, args)
		})
	}
}

func validateArgs(t *testing.T, container *dockerx.Container, mode dockerx.RunMode, args []string) {
	assert := require.New(t)

	switch mode {
	case dockerx.Interactive:
		assert.Contains(args, "-ti")
		assert.NotContains(args, "-d")
	case dockerx.Deamon:
		assert.NotContains(args, "-ti")
		assert.Contains(args, "-d")
	}

	for _, env := range container.Env {
		assert.True(isSubsequence(args, "--env", env))
	}

	if container.Hostname != "" {
		assert.True(isSubsequence(args, dockerx.Hostname...), "[%v] is not a subsequence of [%v]", dockerx.Hostname, args)
	}

	assert.Contains(args, dockerx.ImageName[0])
}

func isSubsequence(list []string, subsequence ...string) bool {
	if len(subsequence) == 0 {
		return true
	}

	for i, val := range list {
		if val == subsequence[0] && reflect.DeepEqual(list[i:i+len(subsequence)], subsequence) {
			return true
		}
	}

	return false
}
