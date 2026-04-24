package ds

import (
	"context"

	dsr3 "github.com/aserto-dev/topaz/api/directory/v4/reader"
	"github.com/aserto-dev/topaz/azm/cache"
	"github.com/aserto-dev/topaz/azm/safe"

	bolt "go.etcd.io/bbolt"
)

type getGraph struct {
	*safe.SafeGetGraph
}

func GetGraph(i *dsr3.GraphRequest) *getGraph {
	return &getGraph{safe.GetGraph(i)}
}

func (i *getGraph) Exec(ctx context.Context, tx *bolt.Tx, mc *cache.Cache) (*dsr3.GraphResponse, error) {
	return mc.GetGraph(i.GraphRequest, getRelations(ctx, tx))
}
