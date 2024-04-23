package edgesync

import (
	"context"

	"github.com/aserto-dev/go-aserto/client"
	"github.com/aserto-dev/go-directory/pkg/datasync"
	"github.com/aserto-dev/go-edge-ds/pkg/directory"

	"github.com/aserto-dev/topaz/resolvers"
	"github.com/rs/zerolog"
)

type Resolver struct {
	logger *zerolog.Logger
	cfg    *client.Config
}

var _ resolvers.EdgeSyncResolver = &Resolver{}

func NewResolver(logger *zerolog.Logger, cfg *client.Config) resolvers.EdgeSyncResolver {
	return &Resolver{
		logger: logger,
		cfg:    cfg,
	}
}

// GetDataSync - returns a data sync client, directly connected to the importer service.
func (r *Resolver) GetDataSync(ctx context.Context) (datasync.Client, error) {
	ds, err := directory.Get()
	if err != nil {
		return nil, err
	}

	return ds.DataSyncClient(), nil
}
