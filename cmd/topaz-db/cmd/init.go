package cmd

import (
	"context"
	"io"
	"os"
	"time"

	eds "github.com/aserto-dev/topaz/internal/pkg/eds"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/directory"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const requestTimeout = 5 * time.Second

func (cmd *InitCmd) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if fi, err := os.Stat(cmd.DBFile); err == nil {
		if fi.IsDir() {
			return errors.Errorf("%s is a directory", cmd.DBFile)
		}

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
