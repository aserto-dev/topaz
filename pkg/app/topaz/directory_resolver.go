package topaz

import (
	"context"

	"github.com/aserto-dev/topaz/pkg/app/directory"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/rs/zerolog"
)

func DirectoryResolver(
	ctx context.Context,
	logger *zerolog.Logger,
	cfg *config.Config,
) resolvers.DirectoryResolver {
	return directory.NewResolver(logger, &cfg.DirectoryResolver)
}
