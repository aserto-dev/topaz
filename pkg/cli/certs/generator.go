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
			fmt.Fprintf(c.UI.Output(), "\n\nSome cert files already exist, skipping generation.\n")
			for _, fn := range existingFiles {
				fmt.Fprintf(c.UI.Output(), "%s (SKIPPED)\n", fn)
			}
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
