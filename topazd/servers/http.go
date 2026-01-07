package servers

import (
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/go-http-utils/headers"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/samber/lo"
)

type HTTPServer struct {
	config.Listener

	HostedDomain      string        `json:"hosted_domain"`
	AllowedOrigins    []string      `json:"allowed_origins"`
	AllowedHeaders    []string      `json:"allowed_headers"`
	AllowedMethods    []string      `json:"allowed_methods"`
	ReadTimeout       time.Duration `json:"read_timeout"`
	ReadHeaderTimeout time.Duration `json:"read_header_timeout"`
	WriteTimeout      time.Duration `json:"write_timeout"`
	IdleTimeout       time.Duration `json:"idle_timeout"`
}

func (s *HTTPServer) Defaults() map[string]any {
	return map[string]any{
		"listen_address":      "0.0.0.0:8383",
		"allowed_origins":     DefaultAllowedOrigins(s.Certs.HasCert()),
		"allowed_headers":     DefaultAllowedHeaders(),
		"allowed_methods":     DefaultAllowedMethods(),
		"read_timeout":        DefaultReadTimeout.String(),
		"read_header_timeout": DefaultReadHeaderTimeout.String(),
		"write_timeout":       DefaultWriteTimeout.String(),
		"idle_timeout":        DefaultIdleTimeout.String(),
	}
}

func (s *HTTPServer) Validate() error {
	return nil
}

func (s *HTTPServer) BaseURL() (*url.URL, error) {
	if s.IsEmptyAddress() {
		return nil, errors.Wrap(config.ErrConfig, "no listen address")
	}

	if s.HostedDomain != "" {
		return url.Parse(s.HostedDomain)
	}

	addr, port, err := net.SplitHostPort(s.ListenAddress)
	if err != nil {
		return nil, err
	}

	if addr == "0.0.0.0" {
		addr = "localhost"
	}

	return &url.URL{
		Scheme: lo.Ternary(s.Certs.HasCert(), "https", "http"),
		Host:   addr + ":" + port,
	}, nil
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
		s.HostedDomain == "" &&
		s.Certs == zeroCerts &&
		s.ReadHeaderTimeout == zeroDuration &&
		s.ReadTimeout == zeroDuration &&
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
		}
	}

	return []string{
		"https://localhost",
		"https://localhost:*",
		"https://127.0.0.1",
		"https://127.0.0.1:*",
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
