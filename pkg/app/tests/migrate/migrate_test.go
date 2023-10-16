package migrate_test

import (
	"context"
	"io"
	"testing"

	dsc3 "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dse3 "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	atesting "github.com/aserto-dev/topaz/pkg/testing"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrate003(t *testing.T) {
	harness := atesting.SetupOnline(t, func(cfg *config.Config) {
		cfg.Edge.DBPath = atesting.AssetAcmeDBFilePath()
	})
	defer harness.Cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := harness.CreateDirectoryClient(ctx)
	token := ""

	objectCount := 0
	for {
		resp, err := client.Reader.GetObjects(ctx, &dsr3.GetObjectsRequest{Page: &dsc3.PaginationRequest{Size: 100, Token: token}})
		if err != nil {
			t.Fail()
		}
		objectCount += len(resp.Results)
		if resp.Page.NextToken == "" {
			break
		}
		token = resp.Page.NextToken
	}
	t.Logf("objects:   %d", objectCount)

	relationCount := 0

	stream, err := client.Exporter.Export(ctx, &dse3.ExportRequest{
		Options:   uint32(dse3.Option_OPTION_DATA_RELATIONS),
		StartFrom: &timestamppb.Timestamp{},
	})
	require.NoError(t, err)
	require.NotNil(t, stream)

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			require.NoError(t, err)
		}

		_, ok := msg.Msg.(*dse3.ExportResponse_Relation)
		if !ok {
			t.Log("unknown message type, skipped")
			continue
		}

		relationCount++
	}
	t.Logf("relations: %d", relationCount)

	assert.Equal(t, objectCount, 1146)
	assert.Equal(t, relationCount, 1939)
}
