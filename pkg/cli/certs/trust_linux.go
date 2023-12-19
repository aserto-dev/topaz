package certs

import (
	"os"

	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

const CaCertsDir = "/usr/local/share/ca-certificates/aserto/"

func AddTrustedCert(certPath string) error {
	if !dirExists(CaCertsDir) {
		if err := os.MkdirAll(CaCertsDir, 0755); err != nil {
			return errors.Wrap(err, "unable to create ca-certificates directory")
		}
	}

	if err := sh.RunV("cp", certPath, CaCertsDir); err != nil {
		return errors.Wrap(err, "unable to copy ca certificate")
	}

	return updateCaCerts()
}

func RemoveTrustedCert(certPath string, filter string) error {
	if !dirExists(CaCertsDir) {
		// Nothing to remove
		return nil
	}

	if err := sh.RunV("rm", "-rf", CaCertsDir); err != nil {
		return errors.Wrap(err, "unable to remove cert")
	}

	return updateCaCerts()
}

func updateCaCerts() error {
	if err := sh.RunV("sudo", "update-ca-certificates"); err != nil {
		return errors.Wrap(err, "unable to update system ca certificates")
	}

	return nil
}

func dirExists(dirPath string) bool {
	if fi, err := os.Stat(dirPath); err == nil && fi.IsDir() {
		return true
	}
	return false
}
