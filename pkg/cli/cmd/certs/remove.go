package certs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/certs"
)

type RemoveCertFileCmd struct {
	CertsDir string `flag:"" required:"false" default:"${topaz_certs_dir}" help:"path to dev certs folder" `
}

func (cmd *RemoveCertFileCmd) Run(c *cc.CommonCtx) error {
	certsDir := cmd.CertsDir
	if fi, err := os.Stat(certsDir); os.IsNotExist(err) || !fi.IsDir() {
		return fmt.Errorf("directory %s not found", certsDir)
	}

	c.UI.Normal().Msgf("certs directory: %s", certsDir)

	table := c.UI.Normal().WithTable("File", "Action")

	// remove cert from trust store, before delete cert file
	for _, fqn := range getFileList(certsDir, withCACerts()) {
		if fi, err := os.Stat(fqn); os.IsNotExist(err) || fi.IsDir() {
			continue
		}

		fn := filepath.Base(fqn)
		cn := fmt.Sprintf("%s-%s", certCommonName, strings.TrimSuffix(fn, filepath.Ext(fn)))

		table.WithTableRow(fn, "removed from trust store")
		if err := certs.RemoveTrustedCert(fqn, cn); err != nil {
			return err
		}
	}

	for _, fqn := range getFileList(certsDir, withAll()) {
		if fi, err := os.Stat(fqn); os.IsNotExist(err) || fi.IsDir() {
			continue
		}
		if err := os.Remove(fqn); err != nil {
			return err
		}
		table.WithTableRow(filepath.Base(fqn), "deleted")
	}

	table.Do()

	return nil
}
