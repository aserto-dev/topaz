package cmd

import (
	"context"
	"os"
	"strings"

	eds "github.com/aserto-dev/topaz/internal/pkg/eds"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/datasync"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/directory"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func (cmd *SyncCmd) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	opts := []datasync.Option{}

	// validate modes
	for _, m := range cmd.Mode {
		mode := datasync.StrToMode(strings.ToUpper(m))
		if mode == datasync.Unknown {
			return errors.Errorf("unknown mode: %s", m)
		}

		opts = append(opts, datasync.WithMode(mode))
	}

	// create client conn
	conn, err := cmd.Connect(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	cfg := &directory.Config{
		DBPath:         cmd.DBFile,
		RequestTimeout: requestTimeout,
	}

	logger := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)

	dir, err := eds.New(ctx, cfg, &logger)
	if err != nil {
		return err
	}
	defer dir.Close()

	return dir.DataSyncClient().Sync(ctx, conn, opts...)
}
