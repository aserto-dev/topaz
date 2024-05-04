package certs

import (
	"fmt"
	"path/filepath"
)

type CertsCmd struct {
	List     ListCertsCmd      `cmd:"" help:"list dev certs"`
	Generate GenerateCertsCmd  `cmd:"" help:"generate dev certs"`
	Trust    TrustCertsCmd     `cmd:"" help:"trust/untrust dev certs"`
	Remove   RemoveCertFileCmd `cmd:"" help:"remove dev certs"`
}

const (
	gatewayFileName = "gateway"
	grpcFileName    = "grpc"
	certCommonName  = "topaz"
)

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
