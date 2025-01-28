package ds_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	client "github.com/aserto-dev/go-aserto"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	assets_test "github.com/aserto-dev/topaz/pkg/app/tests/assets"
	tc "github.com/aserto-dev/topaz/pkg/app/tests/common"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/common"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/directory"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/templates"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestDirectory(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	t.Logf("\nTEST CONTAINER IMAGE: %q\n", tc.TestImage())

	req := testcontainers.ContainerRequest{
		Image:        tc.TestImage(),
		ExposedPorts: []string{"9292/tcp"},
		Env: map[string]string{
			"TOPAZ_CERTS_DIR":     "/certs",
			"TOPAZ_DB_DIR":        "/data",
			"TOPAZ_DECISIONS_DIR": "/decisions",
		},
		Files: []testcontainers.ContainerFile{
			{
				Reader:            assets_test.ConfigReader(),
				ContainerFilePath: "/config/config.yaml",
				FileMode:          0x700,
			},
		},
		WaitingFor: wait.ForAll(
			wait.ForExposedPort(),
			wait.ForLog("Starting 0.0.0.0:9292 gRPC server"),
		).WithStartupTimeoutDefault(300 * time.Second),
	}

	topaz, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          false,
	})
	require.NoError(t, err)

	if err := topaz.Start(ctx); err != nil {
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		testcontainers.CleanupContainer(t, topaz)
		cancel()
	})

	grpcAddr, err := tc.MappedAddr(ctx, topaz, "9292")
	require.NoError(t, err)

	t.Run("testDirectory", testDirectory(grpcAddr))
}

func testDirectory(addr string) func(*testing.T) {
	return func(t *testing.T) {
		opts := []client.ConnectionOption{
			client.WithAddr(addr),
			client.WithInsecure(true),
		}

		conn, err := client.NewConnection(opts...)
		require.NoError(t, err)
		t.Cleanup(func() { _ = conn.Close() })

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		t.Run("", installTemplate(addr, "../../../../assets/gdrive.json"))

		tests := []struct {
			name string
			test func(*testing.T)
		}{
			{"TestCheck", testCheck(ctx, dsr3.NewReaderClient(conn))},
			{"TestChecks", testChecks(ctx, dsr3.NewReaderClient(conn))},
		}

		for _, testCase := range tests {
			t.Run(testCase.name, testCase.test)
		}
	}
}

func installTemplate(addr, tmpl string) func(*testing.T) {
	return func(t *testing.T) {
		t.Logf("addr: %s tmpl: %s", addr, tmpl)

		t.Setenv("TOPAZ_NO_COLOR", "true")
		c, err := cc.NewCommonContext(context.Background(), true, filepath.Join(cc.GetTopazDir(), common.CLIConfigurationFile))
		require.NoError(t, err)

		dsConfig := &dsc.Config{
			Host:      addr,
			Insecure:  true,
			Plaintext: false,
			Timeout:   10 * time.Second,
		}

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

func SetContext(k, v string) *structpb.Struct {
	return &structpb.Struct{
		Fields: map[string]*structpb.Value{
			k: structpb.NewStringValue(v),
		},
	}
}
