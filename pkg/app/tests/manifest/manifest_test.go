package manifest_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"path"
	"testing"

	dsm3 "github.com/aserto-dev/go-directory/aserto/directory/model/v3"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	atesting "github.com/aserto-dev/topaz/pkg/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestManifest(t *testing.T) {
	tempDir := t.TempDir()
	harness := atesting.SetupOnline(t, func(cfg *config.Config) {
		cfg.Edge.DBPath = path.Join(tempDir, "test.db")
	})
	defer harness.Cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := harness.CreateDirectoryClient(ctx)

	r, err := os.Open(path.Join(atesting.AssetsDir(), "manifest.yaml"))
	require.NoError(t, err)

	// write manifest to store
	bytesSend, err := setManifest(ctx, client.Model, r)
	require.NoError(t, err)

	w, err := os.Create(path.Join(tempDir, "manifest.new.yaml"))
	require.NoError(t, err)
	defer w.Close()

	// get manifest from store
	metadata, body, err := getManifest(ctx, client.Model)
	require.NoError(t, err)
	assert.NotNil(t, metadata)

	// write manifest to temp manifest file
	bytesRecv, err := io.Copy(w, body)
	assert.NoError(t, err)

	assert.Equal(t, bytesSend, bytesRecv)

	// delete manifest
	if err := deleteManifest(ctx, client.Model); err != nil {
		assert.NoError(t, err)
	}

	// delete deleted manifest should not result in an error
	if err := deleteManifest(ctx, client.Model); err != nil {
		assert.NoError(t, err)
	}

	// getManifest should not fail, but return an empty manifest.
	if metadata, body, err := getManifest(ctx, client.Model); err == nil {
		assert.NoError(t, err)
		assert.NotNil(t, metadata)

		buf := make([]byte, 1024)

		n, err := body.Read(buf)
		assert.Error(t, err, "EOF")
		assert.Equal(t, n, 0)
		assert.Len(t, buf, 1024)
	} else {
		assert.NoError(t, err)
	}
}

func getManifest(ctx context.Context, client dsm3.ModelClient) (*dsm3.Metadata, io.Reader, error) {
	stream, err := client.GetManifest(ctx, &dsm3.GetManifestRequest{Empty: &emptypb.Empty{}})
	if err != nil {
		return nil, nil, err
	}

	var metadata *dsm3.Metadata
	data := bytes.Buffer{}

	bytesRecv := 0
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
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

func setManifest(ctx context.Context, client dsm3.ModelClient, r io.Reader) (int64, error) {
	stream, err := client.SetManifest(ctx)
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

func deleteManifest(ctx context.Context, client dsm3.ModelClient) error {
	_, err := client.DeleteManifest(ctx, &dsm3.DeleteManifestRequest{Empty: &emptypb.Empty{}})
	return err
}
