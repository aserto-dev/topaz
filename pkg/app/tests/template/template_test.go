package template_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/aserto-dev/topaz/pkg/cli/cmd/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/magefile/mage/sh"
)

var addr string

func TestMain(m *testing.M) {
	ctx := context.Background()

	absPath, err := filepath.Abs(filepath.Join(".", "config.yaml"))
	if err != nil {
		os.Exit(99)
	}

	r, err := os.Open(absPath)
	if err != nil {
		os.Exit(99)
	}

	req := testcontainers.ContainerRequest{
		// Image: "ghcr.io/aserto-dev/topaz:latest",
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    "../../../../",
			Dockerfile: "Dockerfile.test",
			BuildArgs: map[string]*string{
				"GOARCH": GoARCH(),
			},
			PrintBuildLog: true,
			KeepImage:     true,
		},
		ExposedPorts: []string{"9292/tcp", "9393/tcp"},
		Env: map[string]string{
			"TOPAZ_CERTS_DIR":     "/certs",
			"TOPAZ_DB_DIR":        "/data",
			"TOPAZ_DECISIONS_DIR": "/decisions",
		},
		Files: []testcontainers.ContainerFile{
			{
				Reader:            r,
				HostFilePath:      absPath,
				ContainerFilePath: "/config/config.yaml",
				FileMode:          0x700,
			},
		},

		WaitingFor: wait.ForExposedPort(),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		os.Exit(99)
	}

	host, err := container.Host(ctx)
	if err != nil {
		os.Exit(99)
	}

	mappedPort, err := container.MappedPort(ctx, "9292")
	if err != nil {
		os.Exit(99)
	}

	addr = fmt.Sprintf("%s:%s", host, mappedPort.Port())

	defer func() { _ = container.Terminate(ctx) }()

	exitVal := m.Run()

	os.Exit(exitVal)
}

var tcs = []string{
	"../../../../assets/api-auth.json",
	"../../../../assets/gdrive.json",
	"../../../../assets/github.json",
	"../../../../assets/multi-tenant.json",
	"../../../../assets/simple-rbac.json",
	"../../../../assets/slack.json",
}

func TestTemplate(t *testing.T) {
	t.Logf("addr: %s", addr)

	for _, tmpl := range tcs {
		absPath, err := filepath.Abs(tmpl)
		require.NoError(t, err)

		t.Logf("template: %s", absPath)

		tmpl, err := templates.GetTemplateFromFile(absPath)
		require.NoError(t, err)

		dirPath := filepath.Dir(absPath)
		t.Logf("dir %s", dirPath)

		manifestFile := filepath.Join(dirPath, tmpl.Assets.Manifest)
		t.Logf("manifestFile: %s", manifestFile)

		idpDataDir := filepath.Dir(filepath.Join(dirPath, tmpl.Assets.IdentityData[0]))
		t.Logf("idp_data: %s", idpDataDir)

		domainDataDir := filepath.Dir(filepath.Join(dirPath, tmpl.Assets.DomainData[0]))
		t.Logf("domain_data: %s", domainDataDir)

		assertionsFile := filepath.Join(dirPath, tmpl.Assets.Assertions[0])
		t.Logf("assertionsFile: %s", assertionsFile)

		decisionsFile := filepath.Join(dirPath, tmpl.Assets.Assertions[1])
		t.Logf("decisionsFile: %s", decisionsFile)

		t.Run(absPath, execTemplate(addr, manifestFile, idpDataDir, domainDataDir, assertionsFile, decisionsFile))
	}
}

func execTemplate(addr, manifestFile, idpDataDir, domainDataDir, assertionsFile, decisionsFile string) func(*testing.T) {
	return func(t *testing.T) {
		cli := topazCLI()
		t.Logf("cmd: %s", cli)

		env := map[string]string{
			"TOPAZ_DIRECTORY_SVC":  addr,
			"TOPAZ_AUTHORIZER_SVC": addr,
			"TOPAZ_INSECURE":       "true",
			"TOPAZ_NO_COLOR":       "true",
		}

		execStep(t, env, cli, []string{"ds", "delete", "manifest", "--force"})
		execStep(t, env, cli, []string{"ds", "set", "manifest", manifestFile})
		execStep(t, env, cli, []string{"ds", "import", "-d", idpDataDir})
		execStep(t, env, cli, []string{"ds", "import", "-d", domainDataDir})
		execStep(t, env, cli, []string{"ds", "test", "exec", assertionsFile, "--summary"})
		execStep(t, env, cli, []string{"az", "test", "exec", decisionsFile, "--summary"})
	}
}

func execStep(t *testing.T, env map[string]string, cmd string, args []string) {
	ran, err := sh.Exec(env, os.Stdout, os.Stderr, cmd, args...)
	assert.True(t, ran)
	assert.NoError(t, err)
}

func topazCLI() string {
	relPath := fmt.Sprintf("../../../../dist/topaz_%s_%s/topaz", runtime.GOOS, *(GoARCH()))
	absPath, err := filepath.Abs(relPath)
	if err != nil {
		return relPath
	}
	return absPath
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
