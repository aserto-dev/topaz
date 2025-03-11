package template_test

import (
	"context"
	"testing"
	"time"

	assets_test "github.com/aserto-dev/topaz/pkg/app/tests/assets"
	tc "github.com/aserto-dev/topaz/pkg/app/tests/common"
	azc "github.com/aserto-dev/topaz/pkg/cli/clients/authorizer"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestTemplates(t *testing.T) {
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

var tcs = []string{
	"../../../../assets/v33/acmecorp.json",
	"../../../../assets/v33/api-auth.json",
	"../../../../assets/v33/citadel.json",
	"../../../../assets/v33/gdrive.json",
	"../../../../assets/v33/github.json",
	"../../../../assets/v33/multi-tenant.json",
	"../../../../assets/v33/peoplefinder.json",
	"../../../../assets/v33/simple-rbac.json",
	"../../../../assets/v33/slack.json",
	"../../../../assets/v33/todo.json",
	"../../../../assets/v33/gdrive-scream.json",
	"../../../../assets/v33/api-gateway.json",
	"../../../../assets/v32/acmecorp.json",
	"../../../../assets/v32/citadel.json",
	"../../../../assets/v32/gdrive.json",
}
