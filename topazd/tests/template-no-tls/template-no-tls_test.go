package template_no_tls_test

import (
	"context"
	"testing"
	"time"

	"github.com/aserto-dev/topaz/internal/fs"
	azc "github.com/aserto-dev/topaz/topaz/clients/authorizer"
	dsc "github.com/aserto-dev/topaz/topaz/clients/directory"
	"github.com/aserto-dev/topaz/topaz/x"
	assets_test "github.com/aserto-dev/topaz/topazd/tests/assets"
	tc "github.com/aserto-dev/topaz/topazd/tests/common"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestTemplatesNoTLS(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())

	t.Logf("\nTEST CONTAINER IMAGE: %q\n", tc.TestImage())

	req := testcontainers.ContainerRequest{
		Image:        tc.TestImage(),
		ExposedPorts: []string{"9292/tcp"},
		Env: map[string]string{
			x.EnvTopazCertsDir:     x.DefCertsDir,
			x.EnvTopazDBDir:        x.DefDBDir,
			x.EnvTopazDecisionsDir: x.DefDecisionsDir,
		},
		Files: []testcontainers.ContainerFile{
			{
				Reader:            assets_test.ConfigNoTLSReader(),
				ContainerFilePath: tc.TestConfigFilePath,
				FileMode:          int64(fs.FileModeOwnerRWX),
			},
		},
		WaitingFor: wait.ForAll(
			wait.ForExposedPort(),
			wait.ForLog("Starting 0.0.0.0:9292 gRPC server"),
		).WithStartupTimeoutDefault(tc.TestStartupTimeout),
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
		t.Run("testTemplatesWithNoTLS", tc.InstallTemplate(ctx, dsConfig, azConfig, tmpl))
	}
}

var tcs = []string{
	"../../../assets/v34/acmecorp.json",
	"../../../assets/v34/peoplefinder.json",

	"../../../assets/v34/citadel.json",
	"../../../assets/v34/api-auth.json",
	"../../../assets/v34/api-gateway.json",
	"../../../assets/v34/gdrive.json",
	"../../../assets/v34/github.json",
	"../../../assets/v34/multi-tenant.json",
	"../../../assets/v34/simple-rbac.json",
	"../../../assets/v34/slack.json",
	"../../../assets/v34/todo.json",
}
