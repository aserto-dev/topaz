package certs

import (
	"os"
	"path/filepath"

	"github.com/aserto-dev/certs"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type CertPaths struct {
	Name string
	Cert string
	CA   string
	Key  string
	Dir  string
}

func (c *CertPaths) FindExisting() []string {
	existing := []string{}
	for _, cert := range []string{c.Cert, c.CA, c.Key} {
		if fi, err := os.Stat(cert); !os.IsNotExist(err) && !fi.IsDir() {
			existing = append(existing, cert)
		}
	}

	return existing
}

func GenerateCerts(c *cc.CommonCtx, force bool, dnsNames []string, certPaths ...*CertPaths) error {
	if !force {
		existingFiles := []string{}
		for _, cert := range certPaths {
			existingFiles = append(existingFiles, cert.FindExisting()...)
		}

		if len(existingFiles) != 0 {
			table := c.UI.Normal().WithTable("File", "Action")
			for _, fqn := range existingFiles {
				table.WithTableRow(filepath.Base(fqn), "skipped, file already exists")
			}
			table.Do()
			return nil
		}
	}

	return generate(c, dnsNames, certPaths...)
}

func generate(c *cc.CommonCtx, dnsNames []string, certPaths ...*CertPaths) error {
	logger := zerolog.Nop()
	generator := certs.NewGenerator(&logger)

	table := c.UI.Normal().WithTable("File", "Action")

	for _, certPaths := range certPaths {
		if err := generator.MakeDevCert(&certs.CertGenConfig{
			CommonName:       certPaths.Name,
			CertKeyPath:      certPaths.Key,
			CertPath:         certPaths.Cert,
			CACertPath:       certPaths.CA,
			DefaultTLSGenDir: certPaths.Dir,
			DNSNames:         dnsNames,
		}); err != nil {
			return errors.Wrap(err, "failed to create dev certs")
		}

		table.WithTableRow(filepath.Base(certPaths.CA), "generated")
		table.WithTableRow(filepath.Base(certPaths.Cert), "generated")
		table.WithTableRow(filepath.Base(certPaths.Key), "generated")
	}

	table.Do()

	return nil
}
