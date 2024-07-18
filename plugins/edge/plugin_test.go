package edge_test

import (
	"testing"

	"github.com/aserto-dev/go-grpc/aserto/api/v2"
	"github.com/aserto-dev/topaz/plugins/edge"
	"github.com/stretchr/testify/assert"
)

func TestFold(t *testing.T) {
	t.Log("TestFold")
	assert.Equal(t, api.SyncMode(0), edge.Fold(api.SyncMode_SYNC_MODE_UNKNOWN))
	assert.Equal(t, api.SyncMode(0), edge.Fold(api.SyncMode_SYNC_MODE_UNKNOWN, api.SyncMode_SYNC_MODE_UNKNOWN))
	assert.Equal(t, api.SyncMode(8), edge.Fold(api.SyncMode_SYNC_MODE_MANIFEST, api.SyncMode_SYNC_MODE_MANIFEST))
	assert.Equal(t, api.SyncMode(1), edge.Fold(api.SyncMode_SYNC_MODE_UNKNOWN, api.SyncMode_SYNC_MODE_FULL))
	assert.Equal(t, api.SyncMode(2), edge.Fold(api.SyncMode_SYNC_MODE_DIFF))
	assert.Equal(t, api.SyncMode(3), edge.Fold(api.SyncMode_SYNC_MODE_FULL, api.SyncMode_SYNC_MODE_DIFF))
	assert.Equal(t, api.SyncMode(7), edge.Fold(api.SyncMode_SYNC_MODE_FULL, api.SyncMode_SYNC_MODE_DIFF, api.SyncMode_SYNC_MODE_WATERMARK))
	assert.Equal(t, api.SyncMode(15), edge.Fold(api.SyncMode_SYNC_MODE_FULL, api.SyncMode_SYNC_MODE_DIFF, api.SyncMode_SYNC_MODE_WATERMARK, api.SyncMode_SYNC_MODE_MANIFEST))
	all1 := edge.Fold(api.SyncMode_SYNC_MODE_FULL, api.SyncMode_SYNC_MODE_DIFF, api.SyncMode_SYNC_MODE_WATERMARK, api.SyncMode_SYNC_MODE_MANIFEST)
	all2 := edge.Fold(api.SyncMode_SYNC_MODE_FULL, api.SyncMode_SYNC_MODE_DIFF, api.SyncMode_SYNC_MODE_WATERMARK, api.SyncMode_SYNC_MODE_MANIFEST)
	assert.Equal(t, api.SyncMode(15), edge.Fold(all1, all2))
	assert.Equal(t, api.SyncMode(4), edge.Fold(api.SyncMode_SYNC_MODE_WATERMARK, api.SyncMode_SYNC_MODE_WATERMARK, api.SyncMode_SYNC_MODE_WATERMARK))

	req := &api.Command_SyncEdgeDirectory{
		SyncEdgeDirectory: &api.SyncEdgeDirectoryCommand{
			Mode: api.SyncMode_SYNC_MODE_UNKNOWN,
		},
	}
	assert.Equal(t, api.SyncMode(0), edge.Fold(req.SyncEdgeDirectory.Mode))

	t.Log(edge.PrintMode(api.SyncMode_SYNC_MODE_UNKNOWN))
}
