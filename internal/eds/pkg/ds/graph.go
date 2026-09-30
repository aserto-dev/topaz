package ds

import (
	"context"

	dsr "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/internal/azm/cache"
	"github.com/aserto-dev/topaz/internal/azm/safe"

	bolt "go.etcd.io/bbolt"
)

type getGraph struct {
	*safe.SafeGetGraph
}

func GetGraph(i *dsr.GetGraphRequest) *getGraph {
	return &getGraph{safe.GetGraph(i)}
}

func (i *getGraph) Exec(ctx context.Context, tx *bolt.Tx, mc *cache.Cache) (*dsr.GetGraphResponse, error) {
	return mc.GetGraph(i.GetGraphRequest, getRelations(ctx, tx))
}
