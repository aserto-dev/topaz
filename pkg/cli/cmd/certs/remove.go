package certs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/certs"
	"github.com/aserto-dev/topaz/pkg/cli/table"
	"github.com/pkg/errors"
)

type RemoveCertFileCmd struct {
	CertsDir string `flag:"" required:"false" default:"${topaz_certs_dir}" help:"path to dev certs folder" `
}

func (cmd *RemoveCertFileCmd) Run(c *cc.CommonCtx) error {
	certsDir := cmd.CertsDir
	if fi, err := os.Stat(certsDir); os.IsNotExist(err) || !fi.IsDir() {
		return errors.Errorf("directory %s not found", certsDir)
	}

	c.Con().Info().Msg("certs directory: %s", certsDir)

	tab := table.New(c.StdOut())
	defer tab.Close()

	tab.Header("File", "Action")

	data := [][]any{}

	// remove cert from trust store, before delete cert file
	for _, fqn := range getFileList(certsDir, withCACerts()) {
		if fi, err := os.Stat(fqn); os.IsNotExist(err) || fi.IsDir() {
			continue
		}

		fn := filepath.Base(fqn)
		cn := fmt.Sprintf("%s-%s", certCommonName, strings.TrimSuffix(fn, filepath.Ext(fn)))

		data = append(data, []any{fn, "removed from trust store"})

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

		data = append(data, []any{filepath.Base(fqn), "deleted"})
	}

	tab.Bulk(data)
	tab.Render()

	return nil
}
