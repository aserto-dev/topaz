package certs

import (
	"fmt"

	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

const certStore = `Cert:\CurrentUser\Root`

func AddTrustedCert(certPath string) error {
	importCmd := fmt.Sprintf("import-certificate -Filepath %s -CertStoreLocation %s", certPath, certStore)

	if err := sh.RunV("pwsh", "-Command", importCmd); err != nil {
		if err := sh.RunV("powershell", "-Command", importCmd); err != nil {
			return errors.Wrap(err, "failed to import dev certificate to "+certStore)
		}
	}

	return nil
}

func RemoveTrustedCert(certPath, filter string) error {
	removeCmd := fmt.Sprintf("get-childitem %s | Where-Object -Property Subject -match '%s' | Remove-Item", certStore, filter)

	if err := sh.RunV("pwsh", "-Command", removeCmd); err != nil {
		if err := sh.RunV("powershell", "-Command", removeCmd); err != nil {
			return errors.Wrap(err, "failed to remove dev certificate from "+certStore)
		}
	}

	return nil
}
