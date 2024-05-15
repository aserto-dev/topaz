package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	eds "github.com/aserto-dev/go-edge-ds"
	"github.com/aserto-dev/go-edge-ds/pkg/datasync"
	"github.com/aserto-dev/go-edge-ds/pkg/directory"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/rs/zerolog"
)

func (cmd *SyncCmd) Run() error {
	opts := []datasync.Option{}

	// validate modes
	for _, m := range cmd.Mode {
		mode := datasync.StrToMode(strings.ToLower(m))
		if mode == datasync.Unknown {
			return fmt.Errorf("unknown mode: %s", m)
		}
		opts = append(opts, datasync.WithMode(mode))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create client conn
	conn, err := clients.NewDirectoryConn(ctx, &cmd.DirectoryConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	cfg := &directory.Config{
		DBPath:         cmd.DBFile,
		RequestTimeout: 5 * time.Second,
	}

	logger := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)

	dir, err := eds.New(ctx, cfg, &logger)
	if err != nil {
		return err
	}
	defer dir.Close()

	return dir.DataSyncClient().Sync(ctx, conn, opts...)
}
