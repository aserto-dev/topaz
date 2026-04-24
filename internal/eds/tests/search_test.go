package tests_test

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	dsr3 "github.com/aserto-dev/topaz/api/directory/v4/reader"
)

func BenchmarkSearchSerial(b *testing.B) {
	assert := require.New(b)

	checks, err := loadChecks[dsr3.GraphRequest]()
	assert.NoError(err)
	assert.NotEmpty(checks)

	client, cleanup := testInit()
	b.Cleanup(cleanup)

	setupBenchmark(b, client)

	ctx := b.Context()

	b.ResetTimer()

	for _, check := range checks {
		_, err := client.V3.Reader.Graph(ctx, check)
		assert.NoError(err)
	}
}

func BenchmarkSearchParallelChunks(b *testing.B) {
	assert := require.New(b)

	checks, err := loadChecks[dsr3.GraphRequest]()
	assert.NoError(err)
	assert.NotEmpty(checks)

	client, cleanup := testInit()
	b.Cleanup(cleanup)

	setupBenchmark(b, client)

	ctx := b.Context()

	var chunks [][]*dsr3.GraphRequest

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
					_, err := client.V3.Reader.Graph(ctx, check)
					assert.NoError(err)
				}
			}
		})
	}
}
