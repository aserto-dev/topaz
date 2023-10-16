package migrate_test

import (
	"context"
	"testing"

	"github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	atesting "github.com/aserto-dev/topaz/pkg/testing"

	"github.com/stretchr/testify/assert"
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
		resp, err := client.Reader.GetObjects(ctx, &reader.GetObjectsRequest{Page: &common.PaginationRequest{Size: 100, Token: token}})
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

	token = ""
	relationCount := 0
	for {
		resp, err := client.Reader.GetRelations(ctx, &reader.GetRelationsRequest{
			Page: &common.PaginationRequest{Size: 100, Token: token},
		})
		if err != nil {
			t.Fail()
		}
		relationCount += len(resp.Results)
		if resp.Page.NextToken == "" {
			break
		}
		token = resp.Page.NextToken
	}
	t.Logf("relations: %d", relationCount)

	assert.Equal(t, objectCount, 1146)
	assert.Equal(t, relationCount, 1939)
}
