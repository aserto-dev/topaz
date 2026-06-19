package directory

import (
	"context"
	"os"
	"path/filepath"

	dse "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	"github.com/aserto-dev/topaz/internal/fs"
	"github.com/aserto-dev/topaz/topaz/cc"
	"github.com/aserto-dev/topaz/topaz/clients"
	dsc "github.com/aserto-dev/topaz/topaz/clients/directory"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	expObjects   string = "obj"
	expRelations string = "rel"
	expAll       string = "all"
)

type ExportCmd struct {
	dsc.Config

	File   string `flag:"file" short:"f" type:"path" help:"path to target export file"`
	Stdout bool   `flag:"stdout" help:"output to stdout" xor:"file,stdout" required:""`
	Export string `flag:"export" short:"x" enum:"obj,rel,all" default:"all" help:"export [obj|rel|all] types" required:""`
}

func (cmd *ExportCmd) Run(ctx context.Context) error {
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

	cc.Con().Info().Msg(">>> exporting data to %q", cmd.File)

	var writer *os.File

	if cmd.Stdout {
		writer = os.Stdout
	} else {
		writer, err = os.Create(cmd.File)
		if err != nil {
			return err
		}
		defer writer.Close()
	}

	return dsClient.ExportToFile(ctx, writer, cmd.opts())
}

func (cmd *ExportCmd) checkPath() error {
	if cmd.File == "" {
		cmd.Stdout = true
		cmd.File = "stdout"
	}

	if cmd.Stdout {
		return nil
	}

	dir := filepath.Dir(cmd.File)

	ok, err := fs.DirExistsEx(dir)
	if ok && err == nil {
		return nil
	}

	if !ok && err == nil {
		return status.Errorf(codes.NotFound, "directory %q not found", dir)
	}

	return status.Error(codes.Unknown, err.Error())
}

func (cmd *ExportCmd) opts() uint32 {
	switch cmd.Export {
	case expObjects:
		return uint32(dse.Option_OPTION_DATA_OBJECTS)
	case expRelations:
		return uint32(dse.Option_OPTION_DATA_RELATIONS)
	case expAll:
		fallthrough
	default:
		return uint32(dse.Option_OPTION_DATA)
	}
}
