package cmd

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/certs"
)

type CertsCmd struct {
	List     ListCertsCmd     `cmd:"" help:"list dev certs"`
	Generate GenerateCertsCmd `cmd:"" help:"generate dev certs"`
	Trust    TrustCertsCmd    `cmd:"" help:"trust dev certs"`
}

const (
	DefaultCertsDir = "~/.config/topaz/certs"

	gatewayCertFileName = "gateway"
	grpcCertFileName    = "grpc"
	certCommonName      = "topaz"
)

type ListCertsCmd struct {
	CertsDir string `arg:"" required:"false" default:"~/.config/topaz/certs" help:"path to dev certs folder" `
}

func (cmd ListCertsCmd) Run(c *cc.CommonCtx) error {
	certsDir := cmd.CertsDir
	if cmd.CertsDir == DefaultCertsDir {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		certsDir = path.Join(home, "/.config/topaz/certs")
	}

	files, err := os.ReadDir(certsDir)
	if err != nil {
		return err
	}
	certDetails := make(map[string]*x509.Certificate)
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".crt") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(certsDir, file.Name()))
		if err != nil {
			return err
		}
		block, _ := pem.Decode(content)
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return err
		}
		certDetails[file.Name()] = cert
	}
	table := c.UI.Normal().WithTable("File", "Not Before", "Not After", "Valid")

	fileNames := make([]string, 0, len(certDetails))
	for k := range certDetails {
		fileNames = append(fileNames, k)
	}

	sort.Strings(fileNames)

	table.WithTableNoAutoWrapText()
	for _, k := range fileNames {
		isValid := true
		if time.Until(certDetails[k].NotAfter) < 0 {
			isValid = false
		}
		table.WithTableRow(k, certDetails[k].NotBefore.Format(time.RFC3339), certDetails[k].NotAfter.Format(time.RFC3339), fmt.Sprintf("%t", isValid))
	}
	table.Do()
	return nil
}

type GenerateCertsCmd struct {
	TrustCert bool     `flag:"" default:"false" help:"trust generated dev cert"`
	CertsDir  string   `flag:"" required:"false" default:"~/.config/topaz/certs" help:"path to dev cert folder" `
	DNSNames  []string `arg:"" required:"false" default:"" help:"array of DNS Names used in certificate generation"`
}

// Generate a pair of gateway and grpc certificates.
func (cmd GenerateCertsCmd) Run(c *cc.CommonCtx) error {
	c.UI.Normal().Msg("Generating gateway and grpc dev certs")
	certsDir := cmd.CertsDir
	if cmd.CertsDir == DefaultCertsDir {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		certsDir = path.Join(home, "/.config/topaz/certs")
	}

	pathGateway := &certs.CertPaths{
		Name: certCommonName,
		Cert: filepath.Join(certsDir, fmt.Sprintf("%s.crt", gatewayCertFileName)),
		CA:   filepath.Join(certsDir, fmt.Sprintf("%s-ca.crt", gatewayCertFileName)),
		Key:  filepath.Join(certsDir, fmt.Sprintf("%s.key", gatewayCertFileName)),
		Dir:  certsDir,
	}
	pathGRPC := &certs.CertPaths{
		Name: certCommonName,
		Cert: filepath.Join(certsDir, fmt.Sprintf("%s.crt", grpcCertFileName)),
		CA:   filepath.Join(certsDir, fmt.Sprintf("%s-ca.crt", grpcCertFileName)),
		Key:  filepath.Join(certsDir, fmt.Sprintf("%s.key", grpcCertFileName)),
		Dir:  certsDir,
	}
	c.UI.Progress("Please wait").Start()
	err := certs.GenerateCerts(c.UI.Output(), c.UI.Err(), cmd.DNSNames, pathGateway, pathGRPC)
	if err != nil {
		return err
	}
	c.UI.Progress("Done").Stop()
	if cmd.TrustCert {
		certTrust := &TrustCertsCmd{CertsDir: certsDir}
		return certTrust.Run(c)
	}
	return nil
}

type TrustCertsCmd struct {
	CertsDir string `arg:"" required:"false" default:"~/.config/topaz/certs" help:"path to certs folder" `
}

func (cmd TrustCertsCmd) Run(c *cc.CommonCtx) error {
	certsDir := cmd.CertsDir
	if cmd.CertsDir == DefaultCertsDir {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		certsDir = path.Join(home, "/.config/topaz/certs")
	}
	files, err := os.ReadDir(certsDir)
	if err != nil {
		return err
	}
	for _, file := range files {
		c.UI.Normal().Msgf("Adding %s to trusted store", file.Name())
		if !file.IsDir() && strings.HasSuffix(file.Name(), "-ca.crt") {
			if err := certs.AddTrustedCert(filepath.Join(certsDir, file.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}
