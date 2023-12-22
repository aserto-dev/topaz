package certs

import (
	"fmt"
	"os"
	"path"

	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

const loginKeyChain = "Library/Keychains/login.keychain-db"

func AddTrustedCert(certPath string) error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	keyChain := path.Join(homedir, loginKeyChain)

	if err := sh.RunV("security", "add-trusted-cert", "-k", keyChain, certPath); err != nil {
		return errors.Wrap(err, "trusting ca cert")
	}

	return nil
}

func RemoveTrustedCert(certPath, filter string) error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	keyChain := path.Join(homedir, loginKeyChain)

	if !findCert(filter, keyChain) {
		fmt.Println("No certificate to remove.")
		return nil
	}

	if _, err := sh.Output("security", "delete-certificate", "-c", filter, "-t", keyChain); err != nil {
		return errors.Wrap(err, "failed to remove certificate from trust store")
	}

	return nil
}

func findCert(name, keyChain string) bool {
	if _, err := sh.Output("security", "find-certificate", "-c", name, keyChain); err != nil {
		return false
	}
	return true
}
