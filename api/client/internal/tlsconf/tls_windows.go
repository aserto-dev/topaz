package tlsconf

import (
	"crypto/x509"
)

func CertPool(caCertPath string) (*x509.CertPool, error) {
	var certPool *x509.CertPool

	if caCertPath == "" {
		return certPool, nil
	}

	return x509.NewCertPool(), nil
}
