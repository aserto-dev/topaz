package dockerx

import (
	"fmt"
	"os"
	"path"

	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

const (
	Topaz             = "topaz"
	DefaultConfigRoot = ".config/topaz"
)

var (
	DockerRun = sh.RunCmd("docker")
	DockerOut = sh.OutCmd("docker")
)

func DockerWith(env map[string]string, args ...string) error {
	return sh.RunWithV(env, "docker", args...)
}

func DockerWithOut(env map[string]string, args ...string) (string, error) {
	return sh.OutputWith(env, "docker", args...)
}

func IsRunning(name string) (bool, error) {
	if name == "" {
		return false, errors.Errorf("instance name not specified")
	}
	str, err := DockerOut("ps", "-q", "-f", fmt.Sprintf("name=%s", name))
	return str != "", err
}

func DefaultRoots() (confRoot string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return path.Join(home, DefaultConfigRoot), nil
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
