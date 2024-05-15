package cmd

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/aserto-dev/clui"
	dsc "github.com/aserto-dev/go-directory-cli/client"
	"github.com/aserto-dev/go-edge-ds/pkg/directory"
	"github.com/aserto-dev/topaz/cmd/topaz-db/pkg/inproc"

	"github.com/rs/zerolog"
)

func (cmd *SetCmd) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := &directory.Config{
		DBPath:         cmd.DBFile,
		RequestTimeout: 5 * time.Second,
	}

	logger := zerolog.New(io.Discard)

	conn, close := inproc.NewServer(ctx, &logger, cfg)
	defer close()

	dsClient, err := dsc.New(conn, clui.NewUI())
	if err != nil {
		return err
	}

	r := os.Stdin
	if cmd.Manifest != "" {
		r, err = os.Open(cmd.Manifest)
		if err != nil {
			return err
		}
	}

	return dsClient.V3.SetManifest(ctx, r)
}
