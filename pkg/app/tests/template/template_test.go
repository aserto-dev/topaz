package template_no_tls_test

import (
	"context"
	"testing"
	"time"

	assets_test "github.com/aserto-dev/topaz/pkg/app/tests/assets"
	tc "github.com/aserto-dev/topaz/pkg/app/tests/common"
	azc "github.com/aserto-dev/topaz/pkg/cli/clients/authorizer"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
	"github.com/aserto-dev/topaz/pkg/cli/x"

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
			x.EnvTopazCertsDir:  x.DefCertsDir,
			x.EnvTopazDBDir:     x.DefDBDir,
			x.EnvTopazDecisions: x.DefDecisionsDir,
		},
		Files: []testcontainers.ContainerFile{
			{
				Reader:            assets_test.ConfigReader(),
				ContainerFilePath: "/config/config.yaml",
				FileMode:          0o700,
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
		t.Run("testTemplatesWithNoTLS", tc.InstallTemplate(dsConfig, azConfig, tmpl))
	}
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
