package ds

import (
	"context"

	"github.com/aserto-dev/azm/cache"
	"github.com/aserto-dev/azm/safe"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"

	bolt "go.etcd.io/bbolt"
)

type getGraph struct {
	*safe.SafeGetGraph
}

func GetGraph(i *dsr3.GetGraphRequest) *getGraph {
	return &getGraph{safe.GetGraph(i)}
}

func (i *getGraph) Exec(ctx context.Context, tx *bolt.Tx, mc *cache.Cache) (*dsr3.GetGraphResponse, error) {
	return mc.GetGraph(i.GetGraphRequest, getRelations(ctx, tx))
}
