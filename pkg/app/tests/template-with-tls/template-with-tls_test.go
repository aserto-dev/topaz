package template_test

import (
	"context"
	"testing"
	"time"

	assets_test "github.com/aserto-dev/topaz/pkg/app/tests/assets"
	tc "github.com/aserto-dev/topaz/pkg/app/tests/common"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	azc "github.com/aserto-dev/topaz/pkg/cli/clients/authorizer"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/certs"
	"github.com/aserto-dev/topaz/pkg/cli/x"
	"github.com/samber/lo"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestTemplates(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	certsDir := generateCerts(ctx, t)

	t.Logf("\nTEST CONTAINER IMAGE: %q\n", tc.TestImage())

	req := testcontainers.ContainerRequest{
		Image:        tc.TestImage(),
		ExposedPorts: []string{"9292/tcp"},
		Env: map[string]string{
			x.EnvTopazCertsDir:  x.DefCertsDir,
			x.EnvTopazDBDir:     x.DefDBDir,
			x.EnvTopazDecisions: x.DefDecisionsDir,
		},
		Files: append(certFiles(certsDir),
			testcontainers.ContainerFile{
				Reader:            assets_test.ConfigWithTLSReader(),
				ContainerFilePath: "/config/config.yaml",
				FileMode:          0o700,
			},
		),
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

	addr, err := tc.MappedAddr(ctx, topaz, "9292")
	require.NoError(t, err)

	dsConfig := &dsc.Config{
		Host:      addr,
		Insecure:  true,
		Plaintext: false,
		Timeout:   10 * time.Second,
	}

	azConfig := &azc.Config{
		Host:      addr,
		Insecure:  true,
		Plaintext: false,
		Timeout:   10 * time.Second,
	}

	for _, tmpl := range tcs {
		t.Run("testTemplate", tc.InstallTemplate(dsConfig, azConfig, tmpl))
	}
}

func generateCerts(ctx context.Context, t *testing.T) string {
	t.Helper()

	certsDir := t.TempDir()

	commonCtx, err := cc.NewCommonContext(ctx, true, "")
	require.NoError(t, err)

	generateCmd := &certs.GenerateCertsCmd{CertsDir: certsDir}
	require.NoError(t,
		generateCmd.Run(commonCtx),
	)

	return certsDir
}

func certFiles(certsDir string) []testcontainers.ContainerFile {
	return lo.Map(
		[]string{"/grpc.key", "/grpc.crt", "/grpc-ca.crt", "/gateway.key", "/gateway.crt", "/gateway-ca.crt"},
		func(file string, _ int) testcontainers.ContainerFile {
			return testcontainers.ContainerFile{
				HostFilePath:      certsDir + file,
				ContainerFilePath: x.DefCertsDir + file,
				FileMode:          0o600,
			}
		},
	)
}

var tcs = []string{
	"../../../../assets/v32/acmecorp.json",
	"../../../../assets/v32/peoplefinder.json",

	"../../../../assets/v32/citadel.json",
	"../../../../assets/v32/api-auth.json",
	// "../../../../assets/v32/api-gateway.json", // v32/api-gateway.json does not exist.
	"../../../../assets/v32/gdrive.json",
	"../../../../assets/v32/github.json",
	"../../../../assets/v32/multi-tenant.json",
	"../../../../assets/v32/simple-rbac.json",
	"../../../../assets/v32/slack.json",
	"../../../../assets/v32/todo.json",

	"../../../../assets/v33/acmecorp.json",
	"../../../../assets/v33/peoplefinder.json",

	"../../../../assets/v33/citadel.json",
	"../../../../assets/v33/api-auth.json",
	"../../../../assets/v33/api-gateway.json",
	"../../../../assets/v33/gdrive.json",
	"../../../../assets/v33/github.json",
	"../../../../assets/v33/multi-tenant.json",
	"../../../../assets/v33/simple-rbac.json",
	"../../../../assets/v33/slack.json",
	"../../../../assets/v33/todo.json",
}
