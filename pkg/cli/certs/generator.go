package certs

import (
	"fmt"
	"io"
	"os"

	"github.com/aserto-dev/certs"
	"github.com/aserto-dev/logger"
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

func GenerateCerts(logOut, errOut io.Writer, certPaths ...*CertPaths) error {
	existingFiles := []string{}

	for _, cert := range certPaths {
		existingFiles = append(existingFiles, cert.FindExisting()...)
	}

	if len(existingFiles) != 0 {
		fmt.Fprintln(logOut, "Some cert files already exist. Skipping generation.", existingFiles)
		return nil
	}

	return generate(logOut, errOut, certPaths...)
}

func generate(logOut, errOut io.Writer, certPaths ...*CertPaths) error {
	zerologLogger, err := logger.NewLogger(
		logOut, errOut,
		&logger.Config{Prod: false, LogLevel: "warn", LogLevelParsed: zerolog.WarnLevel},
	)
	if err != nil {
		return errors.Wrap(err, "failed to create logger")
	}

	generator := certs.NewGenerator(zerologLogger)

	for _, certPaths := range certPaths {
		if err := generator.MakeDevCert(&certs.CertGenConfig{
			CommonName:       certPaths.Name,
			CertKeyPath:      certPaths.Key,
			CertPath:         certPaths.Cert,
			CACertPath:       certPaths.CA,
			DefaultTLSGenDir: certPaths.Dir,
		}); err != nil {
			return errors.Wrap(err, "failed to create dev certs")
		}
	}

	return nil
}
