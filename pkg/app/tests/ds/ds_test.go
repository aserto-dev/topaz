package ds_test

import (
	"context"
	"testing"
	"time"

	azc "github.com/aserto-dev/topaz/pkg/cli/clients/authorizer"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"

	client "github.com/aserto-dev/go-aserto"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	assets_test "github.com/aserto-dev/topaz/pkg/app/tests/assets"
	tc "github.com/aserto-dev/topaz/pkg/app/tests/common"

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

	t.Run("testDirectory", testDirectory(dsConfig, azConfig))
}

func testDirectory(dsConfig *dsc.Config, azConfig *azc.Config) func(*testing.T) {
	return func(t *testing.T) {
		opts := []client.ConnectionOption{
			client.WithAddr(dsConfig.Host),
			client.WithInsecure(true),
		}

		conn, err := client.NewConnection(opts...)
		require.NoError(t, err)
		t.Cleanup(func() { _ = conn.Close() })

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		t.Run("", tc.InstallTemplate(dsConfig, azConfig, "../../../../assets/gdrive.json"))

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
