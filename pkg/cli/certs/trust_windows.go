package certs

import (
	"fmt"

	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

const certStore = `Cert:\CurrentUser\Root`

func AddTrustedCert(certPath string) error {
	importCmd := fmt.Sprintf("import-certificate -Filepath %s -CertStoreLocation %s", certPath, certStore)

	if err := sh.RunV("powershell", "-Command", importCmd); err != nil {
		return errors.Wrap(err, "failed to import dev certificate to "+certStore)
	}

	return nil
}

func RemoveTrustedCert(certPath string) error {
	removeCmd := fmt.Sprintf("get-childitem %s | Where-Object { $_.Subject -match 'authorizer-gateway-ca' } | Remove-Item", certStore)

	if err := sh.RunV("powershell", "-Command", removeCmd); err != nil {
		return errors.Wrap(err, "failed to remove dev certificate from "+certStore)
	}

	return nil
}
