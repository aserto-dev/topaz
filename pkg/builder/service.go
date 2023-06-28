package builder

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/aserto-dev/certs"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type ServiceInterface interface {
	Start(context.Context) error
	Stop(context.Context) error
}

type GRPCRegistrations func(server *grpc.Server)

type HandlerRegistrations func(ctx context.Context, mux *runtime.ServeMux, grpcEndpoint string, opts []grpc.DialOption) error

type Server struct {
	Config        *API
	Server        *grpc.Server
	Listener      net.Listener
	Registrations GRPCRegistrations
	Gateway       Gateway
	Health        *HealthServer
	Started       chan bool
}

type Gateway struct {
	Server *http.Server
	Mux    *http.ServeMux
	Certs  *certs.TLSCredsConfig
}

type API struct {
	Needs []string `json:"needs"`
	GRPC  struct {
		ListenAddress string `json:"listen_address"`
		// Default connection timeout is 120 seconds
		// https://godoc.org/google.golang.org/grpc#ConnectionTimeout
		ConnectionTimeoutSeconds uint32               `json:"connection_timeout_seconds"`
		Certs                    certs.TLSCredsConfig `json:"certs"`
	} `json:"grpc"`
	Gateway struct {
		ListenAddress     string               `json:"listen_address"`
		AllowedOrigins    []string             `json:"allowed_origins"`
		Certs             certs.TLSCredsConfig `json:"certs"`
		HTTP              bool                 `json:"http"`
		ReadTimeout       time.Duration        `json:"read_timeout"`
		ReadHeaderTimeout time.Duration        `json:"read_header_timeout"`
		WriteTimeout      time.Duration        `json:"write_timeout"`
		IdleTimeout       time.Duration        `json:"idle_timeout"`
	} `json:"gateway"`
	Health struct {
		ListenAddress string `json:"listen_address"`
	} `json:"health"`
}
