package certs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/certs"
)

type GenerateCertsCmd struct {
	CertsDir string   `flag:"" default:"${topaz_certs_dir}" help:"path to dev certs folder" `
	Force    bool     `flag:"" default:"false" help:"force generation of dev certs, overwriting existing cert files"`
	Trust    bool     `flag:"" default:"false" help:"add generated certs to trust store"`
	DNSNames []string `flag:"" default:"localhost" help:"list of DNS names used to generate dev certs"`
}

// Generate a pair of gateway and grpc certificates.
func (cmd *GenerateCertsCmd) Run(c *cc.CommonCtx) error {
	certsDir := cmd.CertsDir
	if _, err := os.Stat(certsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(certsDir, 0o755); err != nil {
			return err
		}
	}

	pathGateway := &certs.CertPaths{
		Name: certCommonName + "-gateway",
		Cert: filepath.Join(certsDir, fmt.Sprintf("%s.crt", gatewayFileName)),
		CA:   filepath.Join(certsDir, fmt.Sprintf("%s-ca.crt", gatewayFileName)),
		Key:  filepath.Join(certsDir, fmt.Sprintf("%s.key", gatewayFileName)),
		Dir:  certsDir,
	}

	pathGRPC := &certs.CertPaths{
		Name: certCommonName + "-grpc",
		Cert: filepath.Join(certsDir, fmt.Sprintf("%s.crt", grpcFileName)),
		CA:   filepath.Join(certsDir, fmt.Sprintf("%s-ca.crt", grpcFileName)),
		Key:  filepath.Join(certsDir, fmt.Sprintf("%s.key", grpcFileName)),
		Dir:  certsDir,
	}

	c.Con().Info().Msg("certs directory: %s", certsDir)
	err := certs.GenerateCerts(c, cmd.Force, cmd.DNSNames, pathGateway, pathGRPC)
	if err != nil {
		return err
	}

	if cmd.Trust {
		certTrust := &TrustCertsCmd{CertsDir: certsDir}
		return certTrust.Run(c)
	}

	return nil
}
