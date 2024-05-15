package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	eds "github.com/aserto-dev/go-edge-ds"
	"github.com/aserto-dev/go-edge-ds/pkg/directory"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func (cmd *InitCmd) Run() error {
	if fi, err := os.Stat(cmd.DBFile); err == nil {
		if fi.IsDir() {
			return fmt.Errorf("%s is a directory", cmd.DBFile)
		}
		return fmt.Errorf("%s already exists", cmd.DBFile)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := &directory.Config{
		DBPath:         cmd.DBFile,
		RequestTimeout: 5 * time.Second,
	}

	logger := zerolog.New(io.Discard)

	dir, err := eds.New(ctx, cfg, &logger)
	if err != nil {
		log.Error().Err(err).Str("db_file", cmd.DBFile).Msg("init_cmd")
	}
	defer dir.Close()

	return nil
}
