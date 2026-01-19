package certs

import (
	"fmt"

	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

const (
	certStore  = `Cert:\CurrentUser\Root`
	powershell = `powershell`
	pwsh       = `pwsh`
)

func AddTrustedCert(certPath string) error {
	importCmd := fmt.Sprintf("Import-Certificate -Filepath %s -CertStoreLocation %s", certPath, certStore)

	if _, err := sh.Output(powershell, "-Command", importCmd); err != nil {
		if _, err := sh.Output(pwsh, "-Command", importCmd); err != nil {
			return errors.Wrap(err, "failed to import dev certificate to "+certStore)
		}
	}

	return nil
}

func RemoveTrustedCert(certPath, filter string) error {
	removeCmd := fmt.Sprintf("Get-ChildItem %s | Where-Object -Property Subject -match '%s' | Remove-Item", certStore, filter)

	if _, err := sh.Output(powershell, "-Command", removeCmd); err != nil {
		if _, err := sh.Output(pwsh, "-Command", removeCmd); err != nil {
			return errors.Wrap(err, "failed to remove dev certificate from "+certStore)
		}
	}

	return nil
}
