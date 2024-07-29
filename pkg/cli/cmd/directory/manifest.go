package directory

import (
	"io"
	"os"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	dsc "github.com/aserto-dev/topaz/pkg/cli/clients/directory"
	"github.com/aserto-dev/topaz/pkg/cli/cmd/common"

	"github.com/pkg/errors"
)

type GetManifestCmd struct {
	Path   string `arg:"path" help:"filepath to manifest file" type:"path" optional:""`
	Stdout bool   `flag:"" help:"output manifest to --stdout"`
	dsc.DirectoryConfig
}

type SetManifestCmd struct {
	Path  string `arg:"" help:"filepath to manifest file" type:"path" optional:""`
	Stdin bool   `flag:"" help:"set manifest from --stdin"`
	dsc.DirectoryConfig
}

type DeleteManifestCmd struct {
	Force bool `flag:"" help:"do not ask for conformation to delete manifest"`
	dsc.DirectoryConfig
}

func (cmd *GetManifestCmd) Run(c *cc.CommonCtx) error {
	if !c.IsServing(cmd.Host) {
		return errors.Wrap(cc.ErrNotServing, cmd.Host)
	}
	dsClient, err := dsc.NewClient(c, &cmd.DirectoryConfig)
	if err != nil {
		return err
	}

	c.Con().Info().Msg(">>> get manifest to %s", cmd.Path)

	r, err := dsClient.GetManifest(c.Context)
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

func (cmd *SetManifestCmd) Run(c *cc.CommonCtx) error {
	if !c.IsServing(cmd.Host) {
		return errors.Wrap(cc.ErrNotServing, cmd.Host)
	}
	dsClient, err := dsc.NewClient(c, &cmd.DirectoryConfig)
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

	c.Con().Info().Msg(">>> set manifest to %s\n", cmd.Path)

	return dsClient.SetManifest(c.Context, r)
}

func (cmd *DeleteManifestCmd) Run(c *cc.CommonCtx) error {
	if !c.IsServing(cmd.Host) {
		return errors.Wrap(cc.ErrNotServing, cmd.Host)
	}
	dsClient, err := dsc.NewClient(c, &cmd.DirectoryConfig)
	if err != nil {
		return err
	}

	c.Con().Warn().Msg("WARNING: delete manifest resets all directory state, including relation and object data")
	if cmd.Force || common.PromptYesNo("Do you want to continue?", false) {
		c.Con().Info().Msg(">>> delete manifest")
		return dsClient.DeleteManifest(c.Context)
	}
	return nil
}
