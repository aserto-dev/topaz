//go:build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"

	"github.com/aserto-dev/clui"
	"github.com/aserto-dev/mage-loot/common"
	"github.com/aserto-dev/mage-loot/deps"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

const containerImage string = "topaz"

func init() {
	os.Setenv("GO_VERSION", "1.22")
	os.Setenv("DOCKER_BUILDKIT", "1")
}

// Generate generates all code.
func Generate() error {
	return common.Generate()
}

// Build builds all binaries in ./cmd.
func Build() error {
	return common.BuildReleaser()
}

// BuildAll builds all binaries in ./cmd for
// all configured operating systems and architectures.
func BuildAll() error {
	return common.BuildAllReleaser("--rm-dist", "--snapshot")
}

// Lint runs linting for the entire project.
func Lint() error {
	return common.Lint()
}

// Test runs all tests and generates a code coverage report.
func Test() error {
	testPkgs := []string{
		"github.com/aserto-dev/topaz/pkg/app/tests/authz",
		"github.com/aserto-dev/topaz/pkg/app/tests/builtin",
		"github.com/aserto-dev/topaz/pkg/app/tests/manifest",
		"github.com/aserto-dev/topaz/pkg/app/tests/policy",
		"github.com/aserto-dev/topaz/pkg/app/tests/query",
	}

	for _, testPkg := range testPkgs {
		if err := test(testPkg); err != nil {
			return err
		}
	}

	return nil
}

func test(pkg string, args ...string) error {
	UI := clui.NewUI()
	UI.Normal().Msg("Running tests.")

	return deps.GoDep("gotestsum")(
		append([]string{"--format", "short-verbose", "--"},
			append(args, "-count=1", "-v", pkg, "-coverprofile=cover.out", fmt.Sprintf("-coverpkg=%s", pkg), filepath.Join(common.WorkDir(), "..."))...,
		)...,
	)
}

// DockerImage builds the docker image for the project.
func DockerImage() error {
	err := BuildAll()
	if err != nil {
		return err
	}
	version, err := common.Version()
	if err != nil {
		return errors.Wrap(err, "failed to calculate version")
	}

	return common.DockerImage(fmt.Sprintf("topaz:%s", version))
}

// DockerPush builds the docker image using all tags specified by sver
// and pushes it to the specified registry
func DockerPush(registry, org string) error {
	tags, err := common.DockerTags(registry, fmt.Sprintf("%s/%s", org, containerImage))
	if err != nil {
		return err
	}

	version, err := common.Version()
	if err != nil {
		return errors.Wrap(err, "failed to calculate version")
	}

	for _, tag := range tags {
		common.UI.Normal().WithStringValue("tag", tag).Msg("pushing tag")
		err = common.DockerPush(
			fmt.Sprintf("%s:%s", containerImage, version),
			fmt.Sprintf("%s/%s/%s:%s", registry, org, containerImage, tag),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// Starts GRPC on a local run of Topaz
func GrpcUI() error {
	grpcUiApp := deps.GoDep("grpcui")
	return grpcUiApp(
		"-insecure",
		"127.0.0.1:8282")
}

func Deps() {
	deps.GetAllDeps()
}

// All runs all targets in the appropriate order.
// The targets are run in the following order:
// deps, generate, lint, test, build, dockerImage
func All() error {
	mg.SerialDeps(Deps, Generate, Lint, Test, Build, DockerImage)
	return nil
}

// Release releases the project.
func Release() error {
	if os.Getenv("GITHUB_TOKEN") == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable is undefined")
	}

	if os.Getenv("ASERTO_TAP") == "" {
		return fmt.Errorf("ASERTO_TAP environment variable is undefined")
	}

	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		return fmt.Errorf("GOOGLE_APPLICATION_CREDENTIALS environment variable is undefined")
	}

	if err := writeVersion(); err != nil {
		return err
	}

	return common.Release("--rm-dist")
}

func RunTest() error {
	return sh.RunV("./dist/topazd_"+runtime.GOOS+"_"+runtime.GOARCH+"/topazd", "--config-file", "./pkg/testing/assets/config-local.yaml", "run")
}

func Run() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	topazDir := path.Join(home, ".config/topaz")
	os.Setenv("TOPAZ_DIR", topazDir)
	cfg := path.Join(topazDir, "cfg/config.yaml")

	return sh.RunV("./dist/topazd_"+runtime.GOOS+"_"+runtime.GOARCH+"/topazd", "--config-file", cfg, "run")
}

func writeVersion() error {
	// get tag.
	version, err := exec.Command("git", "describe", "--tags").Output()
	if err != nil {
		return errors.Wrap(err, "failed to get current git tag")
	}

	file, err := os.Create("VERSION.txt")
	if err != nil {
		return errors.Wrap(err, "failed to create version file")
	}

	defer file.Close()

	if _, err := file.Write(version); err != nil {
		return errors.Wrap(err, "failed to write to version file")
	}

	return file.Sync()
}
