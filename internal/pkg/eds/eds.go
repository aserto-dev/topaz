package eds

import (
	"context"

	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/directory"
	"github.com/rs/zerolog"
)

func New(ctx context.Context, config *directory.Config, logger *zerolog.Logger) (*directory.Directory, error) {
	newLogger := logger.With().Str("component", "edge-ds").Logger()

	ds, err := directory.New(ctx, config, &newLogger)
	if err != nil {
		return nil, err
	}

	return ds, nil
}
