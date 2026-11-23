package tlsconf

import (
	"crypto/x509"
)

func CertPool(caCertPath string) (*x509.CertPool, error) {
	if caCertPath == "" {
		return x509.SystemCertPool()
	}

	return x509.NewCertPool(), nil
}
