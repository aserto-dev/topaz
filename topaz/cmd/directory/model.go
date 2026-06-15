package directory

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/aserto-dev/azm/model"
	v3 "github.com/aserto-dev/azm/v3"
	"github.com/aserto-dev/topaz/internal/fs"
	"github.com/aserto-dev/topaz/topaz/clients"
	dsc "github.com/aserto-dev/topaz/topaz/clients/directory"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetModelCmd struct {
	dsc.Config

	File     string `flag:"" help:"file path to model target file" type:"path" optional:""`
	Stdout   bool   `flag:"" help:"output model to --stdout" default:"true"`
	Manifest string `flag:"" help:"file path to manifest source file" type:"existingfile" optional:""`
	Invert   bool   `flag:"" help:"return the inverted model representation" hidden:""`
}

func (cmd *GetModelCmd) Run(ctx context.Context) error {
	model, err := cmd.getModel(ctx)
	if err != nil {
		return err
	}

	if cmd.Invert {
		model = model.Invert()
	}

	buf, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		return err
	}

	var out io.WriteCloser

	switch {
	case cmd.Stdout:
		out = os.Stdout
		defer out.Close()
	case cmd.File != "":
		out, err := os.Create(cmd.File)
		if err != nil {
			return err
		}
		defer out.Close()
	}

	if _, err := out.Write(buf); err != nil {
		return err
	}

	return nil
}

func (cmd *GetModelCmd) getModel(ctx context.Context) (*model.Model, error) {
	switch {
	case (cmd.Manifest != "" && fs.FileExists(cmd.Manifest)):
		return cmd.getModelFromManifest()
	case (cmd.Manifest == ""):
		return cmd.getModelFromService(ctx)
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid or missing argument(s)")
	}
}

func (cmd *GetModelCmd) getModelFromService(ctx context.Context) (*model.Model, error) {
	if ok, err := clients.Validate(ctx, &cmd.Config); !ok {
		return nil, err
	}

	dsClient, err := dsc.NewClient(ctx, &cmd.Config)
	if err != nil {
		return nil, err
	}

	r, err := dsClient.GetModel(ctx)
	if err != nil {
		return nil, err
	}

	return model.New(r)
}

func (cmd *GetModelCmd) getModelFromManifest() (*model.Model, error) {
	return v3.LoadFile(cmd.Manifest)
}
