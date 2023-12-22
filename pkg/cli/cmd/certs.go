package cmd

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/certs"
)

type CertsCmd struct {
	List     ListCertsCmd      `cmd:"" help:"list dev certs"`
	Generate GenerateCertsCmd  `cmd:"" help:"generate dev certs"`
	Trust    TrustCertsCmd     `cmd:"" help:"trust/untrust dev certs"`
	Remove   RemoveCertFileCmd `cmd:"" help:"remove dev cert file"`
}

const (
	gatewayFileName = "gateway"
	grpcFileName    = "grpc"
	certCommonName  = "topaz"
)

type ListCertsCmd struct {
	CertsDir string `flag:"" required:"false" default:"${topaz_certs_dir}" help:"path to dev certs folder" `
}

func (cmd ListCertsCmd) Run(c *cc.CommonCtx) error {
	certsDir := cmd.CertsDir
	if fi, err := os.Stat(certsDir); os.IsNotExist(err) || !fi.IsDir() {
		return fmt.Errorf("directory %s not found", certsDir)
	}

	c.UI.Normal().Msgf("certs directory: %s", certsDir)

	certDetails := make(map[string]*x509.Certificate)

	for _, fqn := range getFileList(certsDir, withCerts()) {
		if fi, err := os.Stat(fqn); os.IsNotExist(err) || fi.IsDir() {
			continue
		}

		content, err := os.ReadFile(fqn)
		if err != nil {
			return err
		}

		block, _ := pem.Decode(content)

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return err
		}

		_, fn := filepath.Split(fqn)
		certDetails[fn] = cert
	}

	table := c.UI.Normal().WithTable("File", "Not Before", "Not After", "Valid", "CN", "DNS names")

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

		table.WithTableRow(k,
			certDetails[k].NotBefore.Format(time.RFC3339),
			certDetails[k].NotAfter.Format(time.RFC3339),
			fmt.Sprintf("%t", isValid),
			certDetails[k].Issuer.CommonName,
			strings.Join(certDetails[k].DNSNames, ","),
		)
	}
	table.Do()

	return nil
}

type GenerateCertsCmd struct {
	CertsDir string   `flag:"" default:"${topaz_certs_dir}" help:"path to dev certs folder" `
	Force    bool     `flag:"" default:"false" help:"force generation of dev certs, overwriting existing cert files"`
	Trust    bool     `flag:"" default:"false" help:"add generated certs to trust store"`
	DNSNames []string `flag:"" default:"localhost" help:"list of DNS names used to generate dev certs"`
}

// Generate a pair of gateway and grpc certificates.
func (cmd GenerateCertsCmd) Run(c *cc.CommonCtx) error {
	certsDir := cmd.CertsDir
	if _, err := os.Stat(certsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(certsDir, 0755); err != nil {
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

	c.UI.Normal().Msgf("certs directory: %s", certsDir)

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

type TrustCertsCmd struct {
	CertsDir string `flag:"" required:"false" default:"${topaz_certs_dir}" help:"path to dev certs folder" `
	Remove   bool   `flag:"" default:"false" help:"remove dev cert from trust store"`
}

func (cmd TrustCertsCmd) Run(c *cc.CommonCtx) error {
	certsDir := cmd.CertsDir
	if fi, err := os.Stat(certsDir); os.IsNotExist(err) || !fi.IsDir() {
		return fmt.Errorf("directory %s not found", certsDir)
	}

	c.UI.Normal().Msgf("certs directory: %s", certsDir)

	table := c.UI.Normal().WithTable("File", "Action")
	defer table.Do()

	list := getFileList(certsDir, withCACerts())
	if len(list) == 0 {
		table.WithTableRow("no files found", "no actions performed")
		return nil
	}

	for _, fqn := range list {
		if fi, err := os.Stat(fqn); os.IsNotExist(err) || fi.IsDir() {
			continue
		}

		if cmd.Remove {
			fn := cc.FileName(fqn)
			cn := fmt.Sprintf("%s-%s", certCommonName, strings.TrimSuffix(fn, filepath.Ext(fn)))

			table.WithTableRow(fn, "removed from trust store")
			if err := certs.RemoveTrustedCert(fqn, cn); err != nil {
				return err
			}
			continue
		}

		if !cmd.Remove {
			table.WithTableRow(cc.FileName(fqn), "added to trust store")
			if err := certs.AddTrustedCert(fqn); err != nil {
				return err
			}
			continue
		}
	}

	return nil
}

type RemoveCertFileCmd struct {
	CertsDir string `flag:"" required:"false" default:"${topaz_certs_dir}" help:"path to dev certs folder" `
}

func (cmd RemoveCertFileCmd) Run(c *cc.CommonCtx) error {
	certsDir := cmd.CertsDir
	if fi, err := os.Stat(certsDir); os.IsNotExist(err) || !fi.IsDir() {
		return fmt.Errorf("directory %s not found", certsDir)
	}

	c.UI.Normal().Msgf("certs directory: %s", certsDir)

	table := c.UI.Normal().WithTable("File", "Action")

	for _, fqn := range getFileList(certsDir, withAll()) {
		if fi, err := os.Stat(fqn); os.IsNotExist(err) || fi.IsDir() {
			continue
		}
		if err := os.Remove(fqn); err != nil {
			return err
		}
		table.WithTableRow(cc.FileName(fqn), "removed")
	}

	table.Do()

	return nil
}

type fileListArgs struct {
	inclCACerts bool
	inclCerts   bool
	inclKeys    bool
}

type fileListOption func(*fileListArgs)

func withAll() fileListOption {
	return func(arg *fileListArgs) {
		arg.inclCACerts = true
		arg.inclCerts = true
		arg.inclKeys = true
	}
}

func withCerts() fileListOption {
	return func(arg *fileListArgs) {
		arg.inclCerts = true
		arg.inclCACerts = true
	}
}

func withCACerts() fileListOption {
	return func(arg *fileListArgs) {
		arg.inclCACerts = true
	}
}

func getFileList(certsDir string, opts ...fileListOption) []string {
	args := &fileListArgs{}
	for _, opt := range opts {
		opt(args)
	}

	filePaths := []string{}

	if args.inclCACerts {
		filePaths = append(filePaths,
			filepath.Join(certsDir, fmt.Sprintf("%s-ca.crt", grpcFileName)),
			filepath.Join(certsDir, fmt.Sprintf("%s-ca.crt", gatewayFileName)),
		)
	}

	if args.inclCerts {
		filePaths = append(filePaths,
			filepath.Join(certsDir, fmt.Sprintf("%s.crt", grpcFileName)),
			filepath.Join(certsDir, fmt.Sprintf("%s.crt", gatewayFileName)),
		)
	}

	if args.inclKeys {
		filePaths = append(filePaths,
			filepath.Join(certsDir, fmt.Sprintf("%s.key", grpcFileName)),
			filepath.Join(certsDir, fmt.Sprintf("%s.key", gatewayFileName)),
		)
	}

	return filePaths
}
