package certs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/certs"
	"github.com/aserto-dev/topaz/pkg/cli/table"
	"github.com/aserto-dev/topaz/pkg/fs"
	"github.com/pkg/errors"
)

type RemoveCertFileCmd struct {
	CertsDir string `flag:"" required:"false" default:"${topaz_certs_dir}" help:"path to dev certs folder" `
}

func (cmd *RemoveCertFileCmd) Run(c *cc.CommonCtx) error {
	certsDir := cmd.CertsDir
	if !fs.DirExists(certsDir) {
		return errors.Errorf("directory %q not found", certsDir)
	}

	c.Con().Info().Msg("certs directory: %q", certsDir)

	tab := table.New(c.StdOut()).WithColumns("File", "Action")

	// remove cert from trust store, before delete cert file
	for _, fqn := range getFileList(certsDir, withCACerts()) {
		if !fs.FileExists(fqn) {
			continue
		}

		fn := filepath.Base(fqn)
		cn := fmt.Sprintf("%s-%s", certCommonName, strings.TrimSuffix(fn, filepath.Ext(fn)))

		tab.WithRow(fn, "removed from trust store")
		if err := certs.RemoveTrustedCert(fqn, cn); err != nil {
			return err
		}
	}

	for _, fqn := range getFileList(certsDir, withAll()) {
		if !fs.FileExists(fqn) {
			continue
		}
		if err := os.Remove(fqn); err != nil {
			return err
		}
		tab.WithRow(filepath.Base(fqn), "deleted")
	}

	tab.Do()

	return nil
}
