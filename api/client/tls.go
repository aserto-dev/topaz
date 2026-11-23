package client

import (
	"crypto/tls"
	"os"

	"github.com/pkg/errors"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/aserto-dev/topaz/api/client/internal/tlsconf"
)

// TLSConfig contains paths to an X509 certificate's key-pair and CA files.
// It can be used to create client or server tls.Config or grpc TransportCredentials.
type TLSConfig struct {
	Cert string `json:"tls_cert_path"`
	Key  string `json:"tls_key_path"`
	CA   string `json:"tls_ca_cert_path"`
}

func (c *TLSConfig) HasCert() bool {
	return c != nil && c.Cert != "" && c.Key != ""
}

func (c *TLSConfig) HasCA() bool {
	return c != nil && c.CA != ""
}

// ServerConfig returns TLS configuration for a server.
func (c *TLSConfig) ServerConfig() (*tls.Config, error) {
	cfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if !c.HasCert() {
		return cfg, nil
	}

	certificate, err := tls.LoadX509KeyPair(c.Cert, c.Key)
	if err != nil {
		return cfg, errors.Wrapf(err, "failed to load server key-pair (%q, %q)", c.Cert, c.Key)
	}

	cfg.Certificates = []tls.Certificate{certificate}

	return cfg, nil
}

// ClientConfig returns TLS configuration for a client.
func (c *TLSConfig) ClientConfig(skipVerify bool) (*tls.Config, error) {
	conf, err := c.ServerConfig()
	if err != nil {
		return &tls.Config{MinVersion: tls.VersionTLS12}, err
	}

	if skipVerify {
		conf.InsecureSkipVerify = true
		return conf, nil
	}

	certPool, err := tlsconf.CertPool(c.CA)
	if err != nil {
		return conf, errors.Wrapf(err, "failed to create certificate pool with ca %q", c.CA)
	}

	if c.HasCA() {
		caCertBytes, err := os.ReadFile(c.CA)
		if err != nil {
			return conf, errors.Wrapf(err, "failed to read ca cert %q", c.CA)
		}

		if !certPool.AppendCertsFromPEM(caCertBytes) {
			return conf, errors.Errorf("failed to append client ca cert %q", c.CA)
		}
	}

	conf.RootCAs = certPool

	return conf, nil
}

// ServerCredentials returns transport credentials for a GRPC server.
func (c *TLSConfig) ServerCredentials() (credentials.TransportCredentials, error) {
	if !c.HasCert() {
		return insecure.NewCredentials(), nil
	}

	tlsConfig, err := c.ServerConfig()
	if err != nil {
		return nil, err
	}

	return credentials.NewTLS(tlsConfig), nil
}

// ClientCredentials returns transport credentials for a GRPC client.
func (c *TLSConfig) ClientCredentials(skipVerify bool) (credentials.TransportCredentials, error) {
	tlsConfig, err := c.ClientConfig(skipVerify)
	if err != nil {
		return nil, err
	}

	return credentials.NewTLS(tlsConfig), nil
}
