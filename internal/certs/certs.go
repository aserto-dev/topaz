package certs

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const fileModeOwnerRW = 0o600

// Generator generates certs without any external dependencies.
type Generator struct {
	logger *zerolog.Logger
}

// CertGenConfig contains details about how cert generation should happen.
type CertGenConfig struct {
	CommonName       string
	CertKeyPath      string
	CertPath         string
	CertCAPath       string
	DefaultTLSGenDir string
	DNSNames         []string
}

// NewGenerator creates a new cert generator.
func NewGenerator(logger *zerolog.Logger) *Generator {
	log := logger.With().Str("component", "cert-generator").Logger()

	return &Generator{
		logger: &log,
	}
}

// MakeDevCert creates a development certificate request and private key.
// It persists it in the work dir and returns the CSR.
func (c *Generator) MakeDevCert(cfg *CertGenConfig) error {
	c.logger.Info().Str("common-name", cfg.CommonName).Str("cert-path", cfg.CertPath).Msg("generating certificate")
	c.logger.Info().Str("common-name", cfg.CommonName).Str("key-path", cfg.CertKeyPath).Msg("generating certificate")
	c.logger.Info().Str("common-name", cfg.CommonName).Str("ca-cert-path", cfg.CertCAPath).Msg("generating certificate")

	gen, err := newGenerator(c.logger, cfg)
	if err != nil {
		return err
	}

	return gen.generate()
}

type generator struct {
	cfg    *CertGenConfig
	logger *zerolog.Logger

	ca   *x509.Certificate
	cert *x509.Certificate
}

const (
	caSerialNumberBits = 128
	certSerialNumber   = 1658
	privateKeyBits     = 4096
	certDirMode        = 0o777
)

func newGenerator(logger *zerolog.Logger, cfg *CertGenConfig) (*generator, error) {
	caMaxSerialNumber := new(big.Int).Lsh(big.NewInt(1), caSerialNumberBits)

	snCA, err := rand.Int(rand.Reader, caMaxSerialNumber)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate serial number")
	}

	ca := &x509.Certificate{
		SerialNumber: snCA,
		Subject: pkix.Name{
			Organization:  []string{"Aserto, Inc."},
			Country:       []string{"US"},
			Province:      []string{"WA"},
			Locality:      []string{"Seattle"},
			StreetAddress: []string{"-"},
			PostalCode:    []string{"-"},
			CommonName:    cfg.CommonName + "-ca",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	ipAddresses := []net.IP{net.IPv4(127, 0, 0, 1), net.IPv4(0, 0, 0, 0), net.IPv6loopback} //nolint:mnd
	dnsNames := []string{}

	for _, h := range getDNSNames(cfg.DNSNames) {
		if ip := net.ParseIP(h); ip != nil {
			ipAddresses = append(ipAddresses, ip)
		} else {
			dnsNames = append(dnsNames, h)
		}
	}

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(certSerialNumber),
		Subject: pkix.Name{
			Organization:  []string{"Aserto, Inc."},
			Country:       []string{"US"},
			Province:      []string{"WA"},
			Locality:      []string{"Seattle"},
			StreetAddress: []string{"-"},
			PostalCode:    []string{"-"},
			CommonName:    cfg.CommonName,
		},
		DNSNames:     dnsNames,
		IPAddresses:  ipAddresses,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	return &generator{cfg: cfg, logger: logger, ca: ca, cert: cert}, nil
}

func (g *generator) generate() error {
	if err := g.checkDir(); err != nil {
		return errors.Wrap(err, "directory verification returned an error")
	}

	g.logger.Debug().Str("file", g.cfg.CertCAPath).Str("common-name", g.cfg.CommonName).Msg("generating ca certificate")

	caPrivKey, err := rsa.GenerateKey(rand.Reader, privateKeyBits)
	if err != nil {
		return err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, g.ca, g.ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return err
	}

	caPEM := new(bytes.Buffer)
	if err := pem.Encode(caPEM, &pem.Block{Type: "CERTIFICATE", Bytes: caBytes}); err != nil {
		return errors.Wrap(err, "failed to encode cert")
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, privateKeyBits)
	if err != nil {
		return err
	}

	g.logger.Info().
		Str("cert-file", g.cfg.CertPath).
		Str("key-file", g.cfg.CertKeyPath).
		Str("common-name", g.cfg.CommonName).
		Msg("signing certificate")

	if err := g.signCert(certPrivKey, caPrivKey); err != nil {
		return err
	}

	if err := os.WriteFile(g.cfg.CertCAPath, caPEM.Bytes(), fileModeOwnerRW); err != nil {
		return errors.Wrap(err, "failed to write ca cert")
	}

	return nil
}

func (g *generator) signCert(certPrivKey, caPrivKey *rsa.PrivateKey) error {
	g.logger.Info().
		Str("cert-file", g.cfg.CertPath).
		Str("key-file", g.cfg.CertKeyPath).
		Str("common-name", g.cfg.CommonName).
		Msg("signing certificate")

	certBytes, err := x509.CreateCertificate(rand.Reader, g.cert, g.ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return err
	}

	certPEM := new(bytes.Buffer)
	if err := pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	}); err != nil {
		return errors.Wrap(err, "failed to encode cert")
	}

	certPrivKeyPEM := new(bytes.Buffer)
	if err := pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	}); err != nil {
		return errors.Wrap(err, "failed to encode key")
	}

	if err := os.WriteFile(g.cfg.CertKeyPath, certPrivKeyPEM.Bytes(), fileModeOwnerRW); err != nil {
		return errors.Wrap(err, "failed to write key")
	}

	if err := os.WriteFile(g.cfg.CertPath, certPEM.Bytes(), fileModeOwnerRW); err != nil {
		return errors.Wrap(err, "failed to write key")
	}

	return nil
}

func getDNSNames(setNames []string) []string {
	// if DNSNames specified use only the specified DNS Names list
	if len(setNames) > 0 {
		return setNames
	}

	dnsNames := []string{"localhost"}

	if hostname, err := os.Hostname(); err == nil {
		// If there's a hostname for the local machine, add it to the cert's DNS names.
		dnsNames = append(dnsNames, hostname)
	}

	return dnsNames
}

func (g *generator) checkDir() error {
	certDir := filepath.Dir(g.cfg.CertPath)
	keyDir := filepath.Dir(g.cfg.CertKeyPath)
	caCertDir := filepath.Dir(g.cfg.CertCAPath)

	if certDir != keyDir || certDir != caCertDir {
		return errors.New("output directory for all configured certificates and keys must be the same")
	}

	if certDir == g.cfg.DefaultTLSGenDir {
		err := os.MkdirAll(certDir, certDirMode)
		if err != nil {
			return errors.Wrapf(err, "failed to create directory '%s'", g.cfg.DefaultTLSGenDir)
		}
	}

	if _, err := os.Stat(certDir); os.IsNotExist(err) {
		return errors.Errorf("output directory '%s' doesn't exist", certDir)
	} else if err != nil {
		return errors.Wrapf(err, "failed to determine if output directory '%s' exists", certDir)
	}

	return nil
}
