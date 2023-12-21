package certs

import (
	"fmt"
	"os"

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
		fi, err := os.Stat(cert)
		if !os.IsNotExist(err) && !fi.IsDir() {
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
			fmt.Fprintln(c.UI.Output(), "Some cert files already exist. Skipping generation.", existingFiles)
			return nil
		}
	}

	return generate(dnsNames, certPaths...)
}

func generate(dnsNames []string, certPaths ...*CertPaths) error {
	logger := zerolog.Nop()
	generator := certs.NewGenerator(&logger)

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
	}

	return nil
}
