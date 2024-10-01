package certs

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/table"
	"github.com/pkg/errors"
)

type ListCertsCmd struct {
	CertsDir string `flag:"" required:"false" default:"${topaz_certs_dir}" help:"path to dev certs folder" `
}

func (cmd *ListCertsCmd) Run(c *cc.CommonCtx) error {
	certsDir := cmd.CertsDir
	if fi, err := os.Stat(certsDir); os.IsNotExist(err) || !fi.IsDir() {
		return errors.Errorf("directory %s not found", certsDir)
	}

	c.Con().Info().Msg("certs directory: %s", certsDir)

	certDetails := make(map[string]*x509.Certificate)

	for _, fqn := range getFileList(certsDir, withCerts()) {
		if fi, err := os.Stat(fqn); os.IsNotExist(err) || fi.IsDir() {
			continue
		}

		content, err := os.ReadFile(fqn)
		if err != nil {
			return err
		}

		block, _ := pem.Decode(content)

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return err
		}

		_, fn := filepath.Split(fqn)
		certDetails[fn] = cert
	}

	tab := table.New(c.StdOut()).WithColumns("File", "Not Before", "Not After", "Valid", "CN", "DNS names")

	fileNames := make([]string, 0, len(certDetails))
	for k := range certDetails {
		fileNames = append(fileNames, k)
	}

	sort.Strings(fileNames)

	tab.WithTableNoAutoWrapText()
	for _, k := range fileNames {
		isValid := true
		if time.Until(certDetails[k].NotAfter) < 0 {
			isValid = false
		}

		tab.WithRow(k,
			certDetails[k].NotBefore.Format(time.RFC3339),
			certDetails[k].NotAfter.Format(time.RFC3339),
			fmt.Sprintf("%t", isValid),
			certDetails[k].Issuer.CommonName,
			strings.Join(certDetails[k].DNSNames, ","),
		)
	}
	tab.Do()

	return nil
}
