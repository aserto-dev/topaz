package tests_test

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"runtime"
	"testing"

	dsi3 "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/server"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func BenchmarkCheckSerial(b *testing.B) {
	assert := require.New(b)

	checks, err := loadChecks[dsr3.CheckRequest]()
	assert.NoError(err)
	assert.NotEmpty(checks)

	client, cleanup := testInit()
	b.Cleanup(cleanup)

	setupBenchmark(b, client)

	ctx := b.Context()

	b.ResetTimer()

	for _, check := range checks {
		_, err := client.V3.Reader.Check(ctx, check)
		assert.NoError(err)
	}
}

func BenchmarkCheckParallel(b *testing.B) {
	assert := require.New(b)

	checks, err := loadChecks[dsr3.CheckRequest]()
	assert.NoError(err)
	assert.NotEmpty(checks)

	client, cleanup := testInit()
	b.Cleanup(cleanup)

	setupBenchmark(b, client)

	ctx := b.Context()

	b.ResetTimer()

	for _, check := range checks {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := client.V3.Reader.Check(ctx, check)
				assert.NoError(err)
			}
		})
	}
}

func BenchmarkCheckParallelChunks(b *testing.B) {
	assert := require.New(b)

	checks, err := loadChecks[dsr3.CheckRequest]()
	assert.NoError(err)
	assert.NotEmpty(checks)

	client, cleanup := testInit()
	b.Cleanup(cleanup)

	setupBenchmark(b, client)

	ctx := b.Context()

	var chunks [][]*dsr3.CheckRequest

	numChunks := runtime.NumCPU()
	chunkSize := (len(checks) + numChunks - 1) / numChunks

	for i := 0; i < len(checks); i += chunkSize {
		end := min(i+chunkSize, len(checks))
		chunks = append(chunks, checks[i:end])
	}

	b.ResetTimer()

	for _, chunk := range chunks {
		b.RunParallel(func(pb *testing.PB) {
			for _, check := range chunk {
				for pb.Next() {
					_, err := client.V3.Reader.Check(ctx, check)
					assert.NoError(err)
				}
			}
		})
	}
}

func setupBenchmark(b *testing.B, client *server.TestEdgeClient) {
	assert := require.New(b)

	manifest, err := os.ReadFile("./data/check/manifest.yaml")
	assert.NoError(err)

	assert.NoError(deleteManifest(client))
	assert.NoError(setManifest(client, manifest))

	g, iCtx := errgroup.WithContext(b.Context())
	stream, err := client.V3.Importer.Import(iCtx)
	assert.NoError(err)

	g.Go(receiver(stream))

	assert.NoError(importFile(stream, "./data/check/objects.json"))
	assert.NoError(importFile(stream, "./data/check/relations.json"))
	assert.NoError(stream.CloseSend())

	assert.NoError(g.Wait())
}

func loadChecks[T any]() ([]*T, error) {
	bin, err := os.ReadFile("./data/check/check.json")
	if err != nil {
		return nil, err
	}

	var checks []*T
	if err := json.Unmarshal(bin, &checks); err != nil {
		return nil, err
	}

	return checks, nil
}

func receiver(stream dsi3.Importer_ImportClient) func() error {
	return func() error {
		for {
			_, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return nil
			}

			if err != nil {
				return err
			}
		}
	}
}
