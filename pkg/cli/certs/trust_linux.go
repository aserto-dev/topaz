package certs

import (
	"github.com/pkg/errors"
)

const devCertsDocLink = `https://www.topaz.sh/docs/command-line-interface/topaz-cli/certs`

//nolint:staticcheck
func AddTrustedCert(certPath string) error {
	const errMsg = `Adding trust to dev cert was requested. 
Trusting the certificate on Linux distributions automatically is not supported. 
For instructions on how to manually trust the dev cert on your Linux distribution, 
go to: %s`

	return errors.Errorf(errMsg, devCertsDocLink)
}

//nolint:staticcheck
func RemoveTrustedCert(certPath, filter string) error {
	const errMsg = `Removing trust from dev cert was requested. 
Trusting the certificate on Linux distributions automatically is not supported. 
For instructions on how to manually trust the dev cert on your Linux distribution, 
go to: %s`

	return errors.Errorf(errMsg, devCertsDocLink)
}
