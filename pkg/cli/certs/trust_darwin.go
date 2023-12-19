package certs

import (
	"fmt"
	"os"
	"path"

	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

const loginKeychain = "Library/Keychains/login.keychain-db"

func AddTrustedCert(certPath string) error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	keychain := path.Join(homedir, loginKeychain)

	if err := sh.RunV("security", "add-trusted-cert", "-k", keychain, certPath); err != nil {
		return errors.Wrap(err, "trusting ca cert")
	}

	return nil
}

func RemoveTrustedCert(certPath string, filter string) error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	keychain := path.Join(homedir, loginKeychain)

	if !findCert(filter, keychain) {
		fmt.Println("No certificate to remove.")
		return nil
	}

	fmt.Fprintln(os.Stderr, "Deleting dev certificate...")
	if err := sh.RunV("security", "delete-certificate", "-c", filter, "-t", keychain); err != nil {
		return errors.Wrap(err, "can't remove ca cert")
	}

	return nil
}

const StatusNotFound = 44

func findCert(name, keychain string) bool {
	if err := sh.RunV("security", "find-certificate", "-c", name, keychain); err != nil {
		return false
	}

	return true
}
