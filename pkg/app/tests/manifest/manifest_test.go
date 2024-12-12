package manifest_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path"
	"testing"
	"time"

	client "github.com/aserto-dev/go-aserto"
	dsc "github.com/aserto-dev/go-aserto/ds/v3"
	dsm3 "github.com/aserto-dev/go-directory/aserto/directory/model/v3"
	assets_test "github.com/aserto-dev/topaz/pkg/app/tests/assets"
	tc "github.com/aserto-dev/topaz/pkg/app/tests/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestManifest(t *testing.T) {
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

	t.Run("testManifest", testManifest(grpcAddr))
}

func testManifest(addr string) func(*testing.T) {
	return func(t *testing.T) {
		opts := []client.ConnectionOption{
			client.WithAddr(addr),
			client.WithInsecure(true),
		}

		dsClient, err := dsc.New(opts...)
		require.NoError(t, err)
		t.Cleanup(func() { _ = dsClient.Close() })

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		// write manifest to store
		bytesSend, err := setManifest(ctx, dsClient.Model, assets_test.ManifestReader())
		require.NoError(t, err)

		tmpDir := t.TempDir()
		w, err := os.Create(path.Join(tmpDir, "manifest.new.yaml"))
		require.NoError(t, err)
		t.Cleanup(func() { _ = w.Close() })

		// get manifest from store
		metadata, body, err := getManifest(ctx, dsClient.Model)
		require.NoError(t, err)
		assert.NotNil(t, metadata)

		// write manifest to temp manifest file
		bytesRecv, err := io.Copy(w, body)
		require.NoError(t, err)

		assert.Equal(t, bytesSend, bytesRecv)

		// delete manifest
		if err := deleteManifest(ctx, dsClient.Model); err != nil {
			assert.NoError(t, err)
		}

		// delete deleted manifest should not result in an error
		if err := deleteManifest(ctx, dsClient.Model); err != nil {
			assert.NoError(t, err)
		}

		// getManifest should not fail, but return an empty manifest.
		if metadata, body, err := getManifest(ctx, dsClient.Model); err == nil {
			assert.NoError(t, err)
			assert.NotNil(t, metadata)

			buf := make([]byte, 1024)

			n, err := body.Read(buf)
			assert.Error(t, err, "EOF")
			assert.Equal(t, 0, n)
			assert.Len(t, buf, 1024)
		} else {
			assert.NoError(t, err)
		}
	}
}

func getManifest(ctx context.Context, dsm dsm3.ModelClient) (*dsm3.Metadata, io.Reader, error) {
	stream, err := dsm.GetManifest(ctx, &dsm3.GetManifestRequest{Empty: &emptypb.Empty{}})
	if err != nil {
		return nil, nil, err
	}

	var metadata *dsm3.Metadata
	data := bytes.Buffer{}

	bytesRecv := 0
	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, nil, err
		}

		if md, ok := resp.GetMsg().(*dsm3.GetManifestResponse_Metadata); ok {
			metadata = md.Metadata
		}

		if body, ok := resp.GetMsg().(*dsm3.GetManifestResponse_Body); ok {
			data.Write(body.Body.Data)
			bytesRecv += len(body.Body.Data)
		}
	}

	return metadata, bytes.NewReader(data.Bytes()), nil
}

const blockSize = 1024 // intentionally small block size for testing chunking

func setManifest(ctx context.Context, dsm dsm3.ModelClient, r io.Reader) (int64, error) {
	stream, err := dsm.SetManifest(ctx)
	if err != nil {
		return 0, err
	}

	bytesSend := int64(0)

	buf := make([]byte, blockSize)
	for {
		n, err := r.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return bytesSend, err
		}
		bytesSend += int64(n)

		if err := stream.Send(&dsm3.SetManifestRequest{
			Msg: &dsm3.SetManifestRequest_Body{
				Body: &dsm3.Body{Data: buf[0:n]},
			},
		}); err != nil {
			return bytesSend, err
		}

		if n < blockSize {
			break
		}
	}

	if _, err := stream.CloseAndRecv(); err != nil {
		return bytesSend, err
	}

	return bytesSend, nil
}

func deleteManifest(ctx context.Context, dsm dsm3.ModelClient) error {
	_, err := dsm.DeleteManifest(ctx, &dsm3.DeleteManifestRequest{Empty: &emptypb.Empty{}})
	return err
}
