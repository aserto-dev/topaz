package certs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/certs"
	"github.com/aserto-dev/topaz/pkg/cli/table"
)

type TrustCertsCmd struct {
	CertsDir string `flag:"" required:"false" default:"${topaz_certs_dir}" help:"path to dev certs folder" `
	Remove   bool   `flag:"" default:"false" help:"remove dev cert from trust store"`
}

func (cmd *TrustCertsCmd) Run(c *cc.CommonCtx) error {
	certsDir := cmd.CertsDir
	if fi, err := os.Stat(certsDir); os.IsNotExist(err) || !fi.IsDir() {
		return fmt.Errorf("directory %s not found", certsDir)
	}

	fmt.Fprintf(c.StdOut(), "certs directory: %s\n", certsDir)

	if runtime.GOOS == `linux` {
		var err error
		if cmd.Remove {
			err = certs.RemoveTrustedCert("", "")
		} else {
			err = certs.AddTrustedCert("")
		}
		fmt.Fprintf(c.StdErr(), "%s\n", err.Error())
		return nil
	}

	t := table.New(c.StdOut()).WithColumns("File", "Action")
	defer t.Do()

	list := getFileList(certsDir, withCACerts())
	if len(list) == 0 {
		t.WithRow("no files found", "no actions performed")
		return nil
	}

	for _, fqn := range list {
		if fi, err := os.Stat(fqn); os.IsNotExist(err) || fi.IsDir() {
			continue
		}

		if cmd.Remove {
			fn := filepath.Base(fqn)
			cn := fmt.Sprintf("%s-%s", certCommonName, strings.TrimSuffix(fn, filepath.Ext(fn)))

			if err := certs.RemoveTrustedCert(fqn, cn); err != nil {
				return err
			}
			t.WithRow(fn, "removed from trust store")
			continue
		}

		if !cmd.Remove {
			if err := certs.AddTrustedCert(fqn); err != nil {
				return err
			}
			t.WithRow(filepath.Base(fqn), "added to trust store")
			continue
		}
	}

	return nil
}
