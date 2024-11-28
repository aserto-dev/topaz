package template_no_tls_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	assets_test "github.com/aserto-dev/topaz/assets"
	tc "github.com/aserto-dev/topaz/pkg/app/tests/common"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	azc "github.com/aserto-dev/topaz/pkg/cli/clients/authorizer"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/authorizer"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/common"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/directory"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/templates"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var addr string

func TestMain(m *testing.M) {
	rc := 0
	defer func() {
		os.Exit(rc)
	}()

	ctx := context.Background()
	h, err := tc.NewHarness(ctx, &testcontainers.ContainerRequest{
		Image:        tc.TestImage(),
		ExposedPorts: []string{"9292/tcp", "9393/tcp", "9494/tcp", "9696/tcp"},
		Env: map[string]string{
			"TOPAZ_CERTS_DIR":     "/certs",
			"TOPAZ_DB_DIR":        "/data",
			"TOPAZ_DECISIONS_DIR": "/decisions",
		},
		Files: []testcontainers.ContainerFile{
			{
				Reader:            assets_test.ConfigNoTLSReader(),
				ContainerFilePath: "/config/config.yaml",
				FileMode:          0x700,
			},
		},
		WaitingFor: wait.ForAll(
			wait.ForExposedPort(),
			wait.ForLog("Starting 0.0.0.0:9393 gateway server"),
		).WithStartupTimeoutDefault(180 * time.Second).WithDeadline(360 * time.Second),
	})
	if err != nil {
		rc = 99
		return
	}

	defer func() {
		if err := h.Close(ctx); err != nil {
			rc = 100
		}
	}()

	addr = h.AddrGRPC(ctx)

	rc = m.Run()
}

var tcs = []string{
	"../../../../assets/acmecorp.json",
	"../../../../assets/api-auth.json",
	"../../../../assets/citadel.json",
	"../../../../assets/gdrive.json",
	"../../../../assets/github.json",
	"../../../../assets/multi-tenant.json",
	"../../../../assets/peoplefinder.json",
	"../../../../assets/simple-rbac.json",
	"../../../../assets/slack.json",
	"../../../../assets/todo.json",
}

func TestTemplateNoTLS(t *testing.T) {
	t.Logf("addr: %s", addr)

	t.Setenv("TOPAZ_NO_COLOR", "true")
	c, err := cc.NewCommonContext(context.Background(), true, filepath.Join(cc.GetTopazDir(), common.CLIConfigurationFile))
	require.NoError(t, err)

	dsConfig := &dsc.Config{
		Host:      addr,
		Insecure:  false,
		Plaintext: true,
		Timeout:   10 * time.Second,
	}

	azConfig := &azc.Config{
		Host:      addr,
		Insecure:  false,
		Plaintext: true,
		Timeout:   10 * time.Second,
	}

	for _, tmpl := range tcs {
		absPath, err := filepath.Abs(tmpl)
		require.NoError(t, err)

		tmpl, err := templates.GetTemplateFromFile(absPath)
		require.NoError(t, err)

		t.Logf("name %s", tmpl.Name)
		t.Logf("template: %s", absPath)

		dirPath := filepath.Dir(absPath)
		t.Logf("dir %s", dirPath)

		manifestFile := filepath.Join(dirPath, tmpl.Assets.Manifest)
		t.Logf("manifestFile: %s", manifestFile)
		t.Run(tmpl.Name+"-DeleteManifest", DeleteManifest(c, dsConfig))
		t.Run(tmpl.Name+"-SetManifest", SetManifest(c, dsConfig, manifestFile))

		if len(tmpl.Assets.IdentityData) > 0 {
			idpDataDir := filepath.Dir(filepath.Join(dirPath, tmpl.Assets.IdentityData[0]))
			t.Logf("idp_data: %s", idpDataDir)
			t.Run(tmpl.Name+"-ImportIdentityData", ImportData(c, dsConfig, idpDataDir))
		}

		if len(tmpl.Assets.DomainData) > 0 {
			domainDataDir := filepath.Dir(filepath.Join(dirPath, tmpl.Assets.DomainData[0]))
			t.Logf("domain_data: %s", domainDataDir)
			t.Run(tmpl.Name+"-ImportDomainData", ImportData(c, dsConfig, domainDataDir))
		}

		if len(tmpl.Assets.Assertions) > 0 {
			assertionsFile := filepath.Join(dirPath, tmpl.Assets.Assertions[0])
			t.Logf("assertionsFile: %s", assertionsFile)
			t.Run(tmpl.Name+"-ExecDirectoryTest", ExecDirectoryTests(c, dsConfig, []string{assertionsFile}))
		}

		if len(tmpl.Assets.Assertions) > 1 {
			decisionsFile := filepath.Join(dirPath, tmpl.Assets.Assertions[1])
			t.Logf("decisionsFile: %s", decisionsFile)
			t.Run(tmpl.Name+"-ExecAuthorizerTest", ExecAuthorizerTests(c, azConfig, []string{decisionsFile}))
		}
	}
}

func DeleteManifest(c *cc.CommonCtx, cfg *dsc.Config) func(*testing.T) {
	return func(t *testing.T) {
		cmd := directory.DeleteManifestCmd{Config: *cfg, Force: true}
		if err := cmd.Run(c); err != nil {
			assert.NoError(t, err)
		}
	}
}

func SetManifest(c *cc.CommonCtx, cfg *dsc.Config, path string) func(*testing.T) {
	return func(t *testing.T) {
		cmd := directory.SetManifestCmd{Config: *cfg, Path: path}
		if err := cmd.Run(c); err != nil {
			assert.NoError(t, err)
		}
	}
}

func ImportData(c *cc.CommonCtx, cfg *dsc.Config, dir string) func(*testing.T) {
	return func(t *testing.T) {
		cmd := directory.ImportCmd{Config: *cfg, Directory: dir}
		if err := cmd.Run(c); err != nil {
			assert.NoError(t, err)
		}
	}
}

func ExecDirectoryTests(c *cc.CommonCtx, cfg *dsc.Config, files []string) func(*testing.T) {
	return func(t *testing.T) {
		cmd := directory.TestExecCmd{Config: *cfg, TestExecCmd: common.TestExecCmd{Files: files, Summary: true, Desc: "on-error"}}
		if err := cmd.Run(c); err != nil {
			assert.NoError(t, err)
		}
	}
}

func ExecAuthorizerTests(c *cc.CommonCtx, cfg *azc.Config, files []string) func(*testing.T) {
	return func(t *testing.T) {
		cmd := authorizer.TestExecCmd{Config: *cfg, TestExecCmd: common.TestExecCmd{Files: files, Summary: true, Desc: "on-error"}}
		if err := cmd.Run(c); err != nil {
			assert.NoError(t, err)
		}
	}
}
