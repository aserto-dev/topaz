package servers

import (
	"time"

	"github.com/aserto-dev/go-aserto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCServer struct {
	ListenAddress     string           `json:"listen_address"`
	FQDN              string           `json:"fqdn"`
	Certs             aserto.TLSConfig `json:"certs"`
	ConnectionTimeout time.Duration    `json:"connection_timeout"` // https://godoc.org/google.golang.org/grpc#ConnectionTimeout
	DisableReflection bool             `json:"disable_reflection"`
}

func (s *GRPCServer) Defaults() map[string]any {
	return map[string]any{
		"listen_address":         "0.0.0:9292",
		"certs.tls_cert_path":    "${TOPAZ_CERTS_DIR}/grpc.crt",
		"certs.tls_key_path":     "${TOPAZ_CERTS_DIR}/grpc.key",
		"certs.tls_ca_cert_path": "${TOPAZ_CERTS_DIR}/grpc-ca.crt",
		"disable_reflection":     false,
	}
}

func (s *GRPCServer) Validate() error {
	return nil
}

func (s *GRPCServer) HasListener() bool {
	return s != nil && s.ListenAddress != ""
}

func (s *GRPCServer) ClientCredentials() (grpc.DialOption, error) {
	if !s.Certs.HasCert() {
		return grpc.WithTransportCredentials(insecure.NewCredentials()), nil
	}

	creds, err := s.Certs.ClientCredentials(true)
	if err != nil {
		return nil, err
	}

	return grpc.WithTransportCredentials(creds), nil
}
