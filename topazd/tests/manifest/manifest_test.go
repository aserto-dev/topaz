package manifest_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	client "github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/go-aserto/ds/v3"
	dsc "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsr "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	dsw "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"

	"github.com/aserto-dev/topaz/internal/fs"
	"github.com/aserto-dev/topaz/topaz/x"
	assets_test "github.com/aserto-dev/topaz/topazd/tests/assets"
	tc "github.com/aserto-dev/topaz/topazd/tests/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestManifest(t *testing.T) {
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
				Reader:            assets_test.ConfigWithTLSReader(),
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

	grpcAddr, err := tc.MappedAddr(ctx, topaz, "9292")
	require.NoError(t, err)

	t.Run("testManifest", testManifest(ctx, grpcAddr))
}

func testManifest(ctx context.Context, addr string) func(*testing.T) {
	return func(t *testing.T) {
		opts := []client.ConnectionOption{
			client.WithAddr(addr),
			client.WithInsecure(true),
		}

		dsClient, err := ds.New(opts...)
		require.NoError(t, err)
		t.Cleanup(func() { _ = dsClient.Close() })

		ctx, cancel := context.WithCancel(ctx)
		t.Cleanup(cancel)

		bytesSend := len(assets_test.ManifestBytes())

		// write manifest to store
		if resp, err := dsClient.Writer.SetManifest(
			ctx,
			&dsw.SetManifestRequest{
				Manifest: &dsc.Manifest{
					Content: assets_test.ManifestBytes(),
				},
			},
		); err == nil {
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Len(t, bytesSend, len(resp.GetManifest().GetContent()))
		} else {
			require.NoError(t, err)
		}

		tmpDir := t.TempDir()
		w, err := os.Create(filepath.Join(tmpDir, "manifest.new.yaml"))
		require.NoError(t, err)
		t.Cleanup(func() { _ = w.Close() })

		// get manifest from store
		manifest, err := dsClient.Reader.GetManifest(ctx, &dsr.GetManifestRequest{})
		require.NoError(t, err)
		assert.NotNil(t, manifest)

		// write manifest to temp manifest file
		bytesRecv, err := w.Write(manifest.GetManifest().GetContent())
		require.NoError(t, err)

		assert.Equal(t, bytesSend, bytesRecv)

		// delete manifest
		if _, err := dsClient.Writer.DeleteManifest(ctx, &dsw.DeleteManifestRequest{}); err != nil {
			assert.NoError(t, err)
		}

		// delete deleted manifest should not result in an error
		if _, err := dsClient.Writer.DeleteManifest(ctx, &dsw.DeleteManifestRequest{}); err != nil {
			assert.NoError(t, err)
		}

		// getManifest should not fail, but return an empty manifest.
		if manifest, err := dsClient.Reader.GetManifest(ctx, &dsr.GetManifestRequest{}); err == nil {
			assert.NoError(t, err)
			manifest.GetManifest().GetContent()
			manifest.GetManifest().GetEtag()
			manifest.GetManifest().GetUpdatedAt()
		} else {
			assert.NoError(t, err)
		}
	}
}
