package templates

import (
	"context"
	"os"
	"path"
	"path/filepath"

	"github.com/aserto-dev/topaz/internal/fs"
	"github.com/aserto-dev/topaz/topaz/cc"
	azc "github.com/aserto-dev/topaz/topaz/clients/authorizer"
	dsc "github.com/aserto-dev/topaz/topaz/clients/directory"
	"github.com/aserto-dev/topaz/topaz/cmd/common"
	"github.com/aserto-dev/topaz/topaz/cmd/directory"
	"github.com/pkg/errors"
)

type ApplyTemplateCmd struct {
	dsc.Config

	Name  string `arg:"" required:"" help:"template name"`
	Force bool   `flag:"" short:"f" default:"false" required:"false" help:"skip confirmation prompt"`
}

func (cmd *ApplyTemplateCmd) Run(ctx context.Context) error {
	templateDir := path.Join(cc.GetTopazTemplateDir(), cmd.Name)
	if !fs.DirExists(templateDir) {
		return errors.Errorf("directory %q does not exist", templateDir)
	}

	if !cmd.Force {
		cc.Con().Warn().Msg("Applying this template will reset the topaz directory content.")

		if !common.PromptYesNo("Do you want to continue?", false) {
			return nil
		}
	}

	if err := cmd.deleteManifest(ctx); err != nil {
		return err
	}

	if err := cmd.setManifest(ctx); err != nil {
		return err
	}

	if err := cmd.importData(ctx); err != nil {
		return err
	}

	if err := cmd.execAssertions(ctx); err != nil {
		return err
	}

	return nil
}

func (cmd *ApplyTemplateCmd) deleteManifest(ctx context.Context) error {
	command := directory.DeleteManifestCmd{
		Force:  true,
		Config: cmd.Config,
	}

	return command.Run(ctx)
}

func (cmd *ApplyTemplateCmd) setManifest(ctx context.Context) error {
	manifestDir := path.Join(cc.GetTopazTemplateDir(), cmd.Name, "model")
	if !fs.DirExists(manifestDir) {
		return errors.Errorf("directory %q does not exist", manifestDir)
	}

	manifest := filepath.Join(manifestDir, "manifest.yaml")
	if !fs.FileExists(manifest) {
		return errors.Errorf("file %q does not exist", manifest)
	}

	command := directory.SetManifestCmd{
		Path:   manifest,
		Config: cmd.Config,
	}

	return command.Run(ctx)
}

func (cmd *ApplyTemplateCmd) importData(ctx context.Context) error {
	dataDir := path.Join(cc.GetTopazTemplateDir(), cmd.Name, "data")
	if !fs.DirExists(dataDir) {
		return errors.Errorf("directory %q does not exist", dataDir)
	}

	command := directory.ImportCmd{
		Directory: dataDir,
		Config:    cmd.Config,
	}

	return command.Run(ctx)
}

func (cmd *ApplyTemplateCmd) execAssertions(ctx context.Context) error {
	assertionsDir := path.Join(cc.GetTopazTemplateDir(), cmd.Name, "assertions")
	if !fs.DirExists(assertionsDir) {
		return errors.Errorf("directory %q does not exist", assertionsDir)
	}

	entries, err := os.ReadDir(assertionsDir)
	if err != nil {
		return err
	}

	tests := []string{}

	for _, v := range entries {
		if v.IsDir() {
			continue
		}

		tests = append(tests, filepath.Join(assertionsDir, v.Name()))
	}

	runner, err := common.NewTestRunner(
		ctx,
		&common.TestExecCmd{
			Files:   tests,
			Stdin:   false,
			Summary: true,
			Format:  "table",
			Desc:    "on-error",
		},
		&azc.Config{
			Host:      cc.AuthorizerSvc(),
			APIKey:    cmd.APIKey,
			Token:     cmd.Token,
			Insecure:  cmd.Insecure,
			Plaintext: cmd.Plaintext,
			Headers:   cmd.Headers,
			Timeout:   cmd.Timeout,
		},
		&dsc.Config{
			Host:      cmd.Host,
			APIKey:    cmd.APIKey,
			Token:     cmd.Token,
			Insecure:  cmd.Insecure,
			Plaintext: cmd.Plaintext,
			Headers:   cmd.Headers,
			Timeout:   cmd.Timeout,
		},
	)
	if err != nil {
		return err
	}

	if err := runner.Run(ctx); err != nil {
		return err
	}

	return nil
}
