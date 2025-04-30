package directory

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	ds "github.com/aserto-dev/go-edge-ds/pkg/directory"

	"github.com/aserto-dev/topaz/pkg/config"
)

type Directory ds.Directory

func NewDirectory(ctx context.Context, cfg *Config) (*Directory, error) {
	if cfg.Store.Provider != BoltDBStorePlugin {
		return nil, errors.Wrap(config.ErrConfig, "only boltdb provider is currently supported")
	}

	dir, err := ds.New(ctx, (*ds.Config)(&cfg.Store.Bolt), zerolog.Ctx(ctx))
	if err != nil {
		return nil, err
	}

	return (*Directory)(dir), nil
}
