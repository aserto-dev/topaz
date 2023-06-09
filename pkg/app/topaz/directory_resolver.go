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
	cfg *config.Config) resolvers.DirectoryResolver {

	dirCfg, err := cfg.Directory.ToRemoteConfig()
	if err != nil {
		logger.Error().Err(err).Msg("cannot configure directory resolver")
	}
	return directory.NewResolver(logger, dirCfg)
}
