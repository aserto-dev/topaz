package directory

import (
	"context"
	"os"

	"github.com/aserto-dev/topaz/internal/fs"
	"github.com/aserto-dev/topaz/topaz/cc"
	"github.com/aserto-dev/topaz/topaz/clients"
	dsc "github.com/aserto-dev/topaz/topaz/clients/directory"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ImportCmd struct {
	dsc.Config

	File  string `flag:"file" short:"f" type:"path" help:"path to source import file"`
	Stdin bool   `flag:"stdin" help:"import data from --stdin" xor:"file,stdin" required:""`
}

func (cmd *ImportCmd) Run(ctx context.Context) error {
	if err := cmd.checkPath(); err != nil {
		return err
	}

	if ok, err := clients.Validate(ctx, &cmd.Config); !ok {
		return err
	}

	dsClient, err := dsc.NewClient(ctx, &cmd.Config)
	if err != nil {
		return err
	}

	cc.Con().Info().Msg(">>> importing data from %q", cmd.File)

	var reader *os.File

	if cmd.Stdin {
		reader = os.Stdin
	} else {
		reader, err = os.Open(cmd.File)
		if err != nil {
			return err
		}
		defer reader.Close()
	}

	return dsClient.ImportFromFile(ctx, reader)
}

func (cmd *ImportCmd) checkPath() error {
	if cmd.File == "" {
		cmd.Stdin = true
		cmd.File = "stdin"
	}

	if cmd.Stdin {
		return nil
	}

	ok, err := fs.FileExistsEx(cmd.File)
	if ok && err == nil {
		return nil
	}

	if !ok && err == nil {
		return status.Errorf(codes.NotFound, "file %q not found", cmd.File)
	}

	return status.Error(codes.Unknown, err.Error())
}
