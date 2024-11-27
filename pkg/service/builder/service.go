package builder

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/aserto-dev/go-aserto"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type ServiceInterface interface {
	Start(context.Context) error
	Stop(context.Context) error
}

type GRPCRegistrations func(server *grpc.Server)

type HandlerRegistrations func(ctx context.Context, mux *runtime.ServeMux, grpcEndpoint string, opts []grpc.DialOption) error

type Service struct {
	Config   *API
	Server   *grpc.Server
	Listener net.Listener
	Gateway  *Gateway
	Started  chan bool
	Cleanup  []func()
}

type Gateway struct {
	Server *http.Server
	Mux    *http.ServeMux
	Certs  *aserto.TLSConfig
}

type API struct {
	Needs []string `json:"needs"`
	GRPC  struct {
		FQDN          string `json:"fqdn"`
		ListenAddress string `json:"listen_address"`
		// Default connection timeout is 120 seconds
		// https://godoc.org/google.golang.org/grpc#ConnectionTimeout
		ConnectionTimeoutSeconds uint32           `json:"connection_timeout_seconds"`
		Certs                    aserto.TLSConfig `json:"certs"`
	} `json:"grpc"`
	Gateway struct {
		FQDN              string           `json:"fqdn"`
		ListenAddress     string           `json:"listen_address"`
		AllowedOrigins    []string         `json:"allowed_origins"`
		AllowedHeaders    []string         `json:"allowed_headers"`
		AllowedMethods    []string         `json:"allowed_methods"`
		Certs             aserto.TLSConfig `json:"certs"`
		HTTP              bool             `json:"http"`
		ReadTimeout       time.Duration    `json:"read_timeout"`
		ReadHeaderTimeout time.Duration    `json:"read_header_timeout"`
		WriteTimeout      time.Duration    `json:"write_timeout"`
		IdleTimeout       time.Duration    `json:"idle_timeout"`
	} `json:"gateway"`
}

func (g *Gateway) AddHandler(pattern string, handler http.HandlerFunc) {
	g.Mux.Handle(pattern, handler)
}
