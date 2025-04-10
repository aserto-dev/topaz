package certs

import (
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
			filepath.Join(certsDir, grpcFileName+"-ca.crt"),
			filepath.Join(certsDir, gatewayFileName+"-ca.crt"),
		)
	}

	if args.inclCerts {
		filePaths = append(filePaths,
			filepath.Join(certsDir, grpcFileName+".crt"),
			filepath.Join(certsDir, gatewayFileName+".crt"),
		)
	}

	if args.inclKeys {
		filePaths = append(filePaths,
			filepath.Join(certsDir, grpcFileName+".key"),
			filepath.Join(certsDir, gatewayFileName+".key"),
		)
	}

	return filePaths
}
