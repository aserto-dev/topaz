package certs

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/certs"
	"github.com/aserto-dev/topaz/pkg/cli/table"
	"github.com/aserto-dev/topaz/pkg/fs"
	"github.com/pkg/errors"
)

type TrustCertsCmd struct {
	CertsDir string `flag:"" required:"false" default:"${topaz_certs_dir}" help:"path to dev certs folder" `
	Remove   bool   `flag:"" default:"false" help:"remove dev cert from trust store"`
}

func (cmd *TrustCertsCmd) Run(c *cc.CommonCtx) error {
	certsDir := cmd.CertsDir
	if !fs.DirExists(certsDir) {
		return errors.Errorf("directory %q not found", certsDir)
	}

	c.Con().Info().Msg("certs directory: %q", certsDir)

	if runtime.GOOS == `linux` {
		var err error
		if cmd.Remove {
			err = certs.RemoveTrustedCert("", "")
		} else {
			err = certs.AddTrustedCert("")
		}
		c.Con().Error().Msg(err.Error())
		return nil
	}

	tab := table.New(c.StdOut()).WithColumns("File", "Action")
	defer tab.Do()

	list := getFileList(certsDir, withCACerts())
	if len(list) == 0 {
		tab.WithRow("no files found", "no actions performed")
		return nil
	}

	for _, fqn := range list {
		if !fs.FileExists(fqn) {
			continue
		}

		if cmd.Remove {
			fn := filepath.Base(fqn)
			cn := fmt.Sprintf("%s-%s", certCommonName, strings.TrimSuffix(fn, filepath.Ext(fn)))

			if err := certs.RemoveTrustedCert(fqn, cn); err != nil {
				return err
			}
			tab.WithRow(fn, "removed from trust store")
			continue
		}

		if !cmd.Remove {
			if err := certs.AddTrustedCert(fqn); err != nil {
				return err
			}
			tab.WithRow(filepath.Base(fqn), "added to trust store")
			continue
		}
	}

	return nil
}
