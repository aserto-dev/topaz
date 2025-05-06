package servers

import (
	"net/http"
	"time"

	"github.com/aserto-dev/go-aserto"
	"github.com/go-http-utils/headers"
	"github.com/rs/cors"
)

type HTTPServer struct {
	ListenAddress     string           `json:"listen_address"`
	FQDN              string           `json:"fqdn"`
	Certs             aserto.TLSConfig `json:"certs"`
	AllowedOrigins    []string         `json:"allowed_origins"`
	AllowedHeaders    []string         `json:"allowed_headers"`
	AllowedMethods    []string         `json:"allowed_methods"`
	ReadTimeout       time.Duration    `json:"read_timeout"`
	ReadHeaderTimeout time.Duration    `json:"read_header_timeout"`
	WriteTimeout      time.Duration    `json:"write_timeout"`
	IdleTimeout       time.Duration    `json:"idle_timeout"`
}

func (s *HTTPServer) Defaults() map[string]any {
	return map[string]any{
		"listen_address":         "0.0.0:9393",
		"certs.tls_cert_path":    "${TOPAZ_CERTS_DIR}/gateway.crt",
		"certs.tls_key_path":     "${TOPAZ_CERTS_DIR}/gateway.key",
		"certs.tls_ca_cert_path": "${TOPAZ_CERTS_DIR}/gateway-ca.crt",
		"allowed_origins":        DefaultAllowedOrigins(s.Certs.HasCert()),
		"allowed_headers":        DefaultAllowedHeaders(),
		"allowed_methods":        DefaultAllowedMethods(),
		"http":                   false,
		"read_timeout":           DefaultReadTimeout.String(),
		"read_header_timeout":    DefaultReadHeaderTimeout.String(),
		"write_timeout":          DefaultWriteTimeout.String(),
		"idle_timeout":           DefaultIdleTimeout.String(),
	}
}

func (s *HTTPServer) Validate() error {
	return nil
}

func (s *HTTPServer) HasListener() bool {
	return s != nil && s.ListenAddress != ""
}

func (s *HTTPServer) Cors() *cors.Cors {
	return cors.New(cors.Options{
		AllowedHeaders: s.AllowedHeaders,
		AllowedOrigins: s.AllowedOrigins,
		AllowedMethods: s.AllowedMethods,
		Debug:          false,
	})
}

func (s *HTTPServer) IsEmpty() bool {
	var (
		zeroCerts    aserto.TLSConfig
		zeroDuration time.Duration
	)

	return s.ListenAddress == "" &&
		s.FQDN == "" &&
		s.Certs == zeroCerts &&
		s.ReadHeaderTimeout == zeroDuration &&
		s.ReadHeaderTimeout == zeroDuration &&
		s.WriteTimeout == zeroDuration &&
		s.IdleTimeout == zeroDuration &&
		len(s.AllowedOrigins) == 0 &&
		len(s.AllowedHeaders) == 0 &&
		len(s.AllowedMethods) == 0
}

const (
	DefaultReadTimeout       = time.Second * 5
	DefaultReadHeaderTimeout = time.Second * 5
	DefaultWriteTimeout      = time.Second * 5
	DefaultIdleTimeout       = time.Second * 30

	listenersPerService = 2 // gRPC and HTTP
)

func DefaultAllowedOrigins(useHTTP bool) []string {
	if useHTTP {
		return []string{
			"http://localhost",
			"http://localhost:*",
			"http://127.0.0.1",
			"http://127.0.0.1:*",
			"http://0.0.0.0",
			"http://0.0.0.0:*",
		}
	}

	return []string{
		"https://localhost",
		"https://localhost:*",
		"https://127.0.0.1",
		"https://127.0.0.1:*",
		"https://0.0.0.0",
		"https://0.0.0.0:*",
	}
}

func DefaultAllowedHeaders() []string {
	return []string{
		headers.Authorization,
		headers.ContentType,
		headers.IfMatch,
		headers.IfNoneMatch,
		"Depth",
	}
}

func DefaultAllowedMethods() []string {
	return []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodHead,
		http.MethodDelete,
		http.MethodPut,
		http.MethodPatch,
		"PROPFIND",
		"MKCOL",
		"COPY",
		"MOVE",
	}
}
