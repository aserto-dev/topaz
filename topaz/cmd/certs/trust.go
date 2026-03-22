package certs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/aserto-dev/topaz/topaz/cc"
	"github.com/aserto-dev/topaz/topaz/certs"
	"github.com/aserto-dev/topaz/topaz/table"
	"github.com/pkg/errors"
)

type TrustCertsCmd struct {
	CertsDir string `flag:"" required:"false" default:"${topaz_certs_dir}" help:"path to dev certs folder" `
	Remove   bool   `flag:"" default:"false" help:"remove dev cert from trust store"`
}

func (cmd *TrustCertsCmd) Run(ctx context.Context) error {
	certsDir := cmd.CertsDir
	if fi, err := os.Stat(certsDir); os.IsNotExist(err) || !fi.IsDir() {
		return errors.Errorf("directory %s not found", certsDir)
	}

	cc.Con().Info().Msg("certs directory: %s", certsDir)

	if runtime.GOOS == `linux` {
		var err error
		if cmd.Remove {
			err = certs.RemoveTrustedCert("", "")
		} else {
			err = certs.AddTrustedCert("")
		}

		cc.Con().Error().Msg(err.Error())

		return nil
	}

	data := [][]any{}

	tab := table.New(os.Stdout)

	defer func() {
		tab.Bulk(data)
		tab.Render()
		tab.Close()
	}()

	tab.Header("File", "Action")

	list := getFileList(certsDir, withCACerts())
	if len(list) == 0 {
		data = append(data, []any{"no files found", "no actions performed"})

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

			data = append(data, []any{fn, "removed from trust store"})

			continue
		}

		if !cmd.Remove {
			if err := certs.AddTrustedCert(fqn); err != nil {
				return err
			}

			data = append(data, []any{filepath.Base(fqn), "added to trust store"})

			continue
		}
	}

	return nil
}
