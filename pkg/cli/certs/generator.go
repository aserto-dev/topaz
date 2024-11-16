package certs

import (
	"os"
	"path/filepath"

	"github.com/aserto-dev/certs"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/table"
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
			tab := table.New(c.StdErr()).WithColumns("File", "Action")
			for _, fqn := range existingFiles {
				tab.WithRow(filepath.Base(fqn), "skipped, file already exists")
			}
			tab.Do()
			return nil
		}
	}

	return generate(c, dnsNames, certPaths...)
}

func generate(c *cc.CommonCtx, dnsNames []string, certPaths ...*CertPaths) error {
	logger := zerolog.Nop()
	generator := certs.NewGenerator(&logger)

	tab := table.New(c.StdErr()).WithColumns("File", "Action")

	for _, certPaths := range certPaths {
		if err := generator.MakeDevCert(&certs.CertGenConfig{
			CommonName:  certPaths.Name,
			CertKeyPath: certPaths.Key,
			CertPath:    certPaths.Cert,
			CertCAPath:  certPaths.CA,
			DNSNames:    dnsNames,
		}); err != nil {
			return errors.Wrap(err, "failed to create dev certs")
		}

		tab.WithRow(filepath.Base(certPaths.CA), "generated")
		tab.WithRow(filepath.Base(certPaths.Cert), "generated")
		tab.WithRow(filepath.Base(certPaths.Key), "generated")
	}

	tab.Do()

	return nil
}
