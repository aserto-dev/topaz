package cmd

import (
	"context"
	"io"
	"time"

	"github.com/aserto-dev/topaz/internal/pkg/eds"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/directory"
	"github.com/aserto-dev/topaz/internal/pkg/fs"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const requestTimeout = 5 * time.Second

func (cmd *InitCmd) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if fs.FileExists(cmd.DBFile) {
		return errors.Errorf("%s already exists", cmd.DBFile)
	}

	cfg := &directory.Config{
		DBPath:         cmd.DBFile,
		RequestTimeout: requestTimeout,
	}

	logger := zerolog.New(io.Discard)

	dir, err := eds.New(ctx, cfg, &logger)
	if err != nil {
		log.Error().Err(err).Str("db_file", cmd.DBFile).Msg("init_cmd")
	}
	defer dir.Close()

	return nil
}
