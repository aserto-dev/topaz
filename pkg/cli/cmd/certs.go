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
	List     ListCertsCmd      `cmd:"" help:"list dev certs"`
	Remove   RemoveCertFileCmd `cmd:"" help:"remove dev cert file"`
	Generate GenerateCertsCmd  `cmd:"" help:"generate dev certs"`
	Trust    TrustCertsCmd     `cmd:"" help:"trust/untrust dev certs"`
}

const (
	DefaultCertsDir = "~/.config/topaz/certs"

	gatewayCertFileName = "gateway"
	grpcCertFileName    = "grpc"
	certCommonName      = "topaz"
)

type ListCertsCmd struct {
	CertsDir string `flag:"" required:"false" default:"~/.config/topaz/certs" help:"path to dev cert folder" `
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
	Force     bool     `flag:"" default:"false" help:"force generate dev cert"`
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
	c.UI.Progress("").Start()
	err := certs.GenerateCerts(c.UI.Output(), c.UI.Err(), cmd.Force, cmd.DNSNames, pathGateway, pathGRPC)
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
	CertsDir string `flag:"" required:"false" default:"~/.config/topaz/certs" help:"path to dev cert folder" `
	Remove   bool   `flag:"" default:"false" help:"remove trusted dev cert"`
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
		if !file.IsDir() && strings.HasSuffix(file.Name(), "-ca.crt") {
			if cmd.Remove {
				c.UI.Normal().Msgf("Removing %s from trusted store", file.Name())
				if err := certs.RemoveTrustedCert(filepath.Join(certsDir, file.Name()), certCommonName); err != nil {
					return err
				}
			} else {
				c.UI.Normal().Msgf("Adding %s to trusted store", file.Name())
				if err := certs.AddTrustedCert(filepath.Join(certsDir, file.Name())); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

type RemoveCertFileCmd struct {
	All          bool   `flag:"" required:"false" default:"false" help:"remove all certs"`
	CertsDir     string `flag:"" required:"false" default:"~/.config/topaz/certs" help:"path to dev cert folder" `
	CertFileName string `arg:"" required:"false" default:"" help:"name of the cert file to remove" `
}

func (cmd RemoveCertFileCmd) Run(c *cc.CommonCtx) error {
	certsDir := cmd.CertsDir
	if cmd.CertsDir == DefaultCertsDir {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		certsDir = path.Join(home, "/.config/topaz/certs")
	}
	if cmd.All {
		c.UI.Progress("Removing all dev certs").Start()
		files, err := os.ReadDir(certsDir)
		if err != nil {
			return err
		}
		for _, file := range files {
			if err := os.Remove(filepath.Join(certsDir, file.Name())); err != nil {
				return err
			}
		}
		c.UI.Progress("Done").Stop()
		return nil
	}

	return os.Remove(filepath.Join(certsDir, cmd.CertFileName))
}
