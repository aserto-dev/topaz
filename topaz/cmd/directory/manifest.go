package directory

import (
	"context"
	"io"
	"os"

	"github.com/aserto-dev/topaz/topaz/cc"
	"github.com/aserto-dev/topaz/topaz/clients"
	dsc "github.com/aserto-dev/topaz/topaz/clients/directory"
	"github.com/aserto-dev/topaz/topaz/cmd/common"
)

type GetManifestCmd struct {
	dsc.Config

	Path   string `arg:"path" help:"filepath to manifest file" type:"path" optional:""`
	Stdout bool   `flag:"" help:"output manifest to --stdout"`
}

type SetManifestCmd struct {
	dsc.Config

	Path  string `arg:"" help:"filepath to manifest file" type:"path" optional:""`
	Stdin bool   `flag:"" help:"set manifest from --stdin"`
}

type DeleteManifestCmd struct {
	dsc.Config

	Force bool `flag:"" help:"do not ask for conformation to delete manifest"`
}

func (cmd *GetManifestCmd) Run(ctx context.Context) error {
	if ok, err := clients.Validate(ctx, &cmd.Config); !ok {
		return err
	}

	dsClient, err := dsc.NewClient(ctx, &cmd.Config)
	if err != nil {
		return err
	}

	cc.Con().Info().Msg(">>> get manifest to %s", cmd.Path)

	r, err := dsClient.GetManifest(ctx)
	if err != nil {
		return err
	}

	w := os.Stdout

	if cmd.Path != "" {
		w, err = os.Create(cmd.Path)
		if err != nil {
			return err
		}
	}

	defer w.Close()

	if _, err := io.Copy(w, r); err != nil {
		return err
	}

	return nil
}

func (cmd *SetManifestCmd) Run(ctx context.Context) error {
	if ok, err := clients.Validate(ctx, &cmd.Config); !ok {
		return err
	}

	dsClient, err := dsc.NewClient(ctx, &cmd.Config)
	if err != nil {
		return err
	}

	r := os.Stdin
	if cmd.Path != "" {
		r, err = os.Open(cmd.Path)
		if err != nil {
			return err
		}
	}

	cc.Con().Info().Msg(">>> set manifest to %s\n", cmd.Path)

	return dsClient.SetManifest(ctx, r)
}

func (cmd *DeleteManifestCmd) Run(ctx context.Context) error {
	if ok, err := clients.Validate(ctx, &cmd.Config); !ok {
		return err
	}

	dsClient, err := dsc.NewClient(ctx, &cmd.Config)
	if err != nil {
		return err
	}

	cc.Con().Warn().Msg("WARNING: delete manifest resets all directory state, including relation and object data")

	if cmd.Force || common.PromptYesNo("Do you want to continue?", false) {
		cc.Con().Info().Msg(">>> delete manifest")
		return dsClient.DeleteManifest(ctx)
	}

	return nil
}
