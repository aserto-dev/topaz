package topaz

import (
	"context"
	"sync"
	"time"

	"github.com/aserto-dev/topaz/api/client"
	"github.com/aserto-dev/topaz/api/directory/v3/reader"
	"github.com/aserto-dev/topaz/topaz-opa/internal/config"
	"github.com/aserto-dev/topaz/topaz-opa/internal/errs"
	"github.com/authzen/access.go/api/access/v1"
	"github.com/open-policy-agent/opa/v1/plugins"

	"google.golang.org/grpc"
)

const PluginName string = `topaz`

var (
	connMu sync.Mutex
	conn   *grpc.ClientConn
)

// getConn returns the shared connection to Topaz, creating it on first use.
func getConn() (*grpc.ClientConn, error) {
	cfg := GetConfig()
	if !cfg.Enabled {
		return nil, errs.ErrTopazPluginDisabled
	}

	connMu.Lock()
	defer connMu.Unlock()

	if conn != nil {
		return conn, nil
	}

	c, err := cfg.Connection.Connect(client.WithChainUnaryInterceptor(timeoutInterceptor(cfg.RequestTimeout)))
	if err != nil {
		return nil, err
	}

	conn = c

	return conn, nil
}

// closeConn closes the shared connection, the next call to getConn reconnects.
func closeConn() {
	connMu.Lock()
	defer connMu.Unlock()

	if conn != nil {
		_ = conn.Close()
		conn = nil
	}
}

// timeoutInterceptor applies the request timeout to every unary call.
func timeoutInterceptor(timeout config.Duration) grpc.UnaryClientInterceptor {
	d := time.Duration(timeout)
	if d <= 0 {
		d = defRequestTimeout
	}

	return func(
		ctx context.Context, method string, req, reply any,
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
	) error {
		ctx, cancel := context.WithTimeout(ctx, d)
		defer cancel()

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

type Plugin struct {
	manager *plugins.Manager
	config  *Config
}

var _ plugins.Plugin = (*Plugin)(nil)

func (p *Plugin) Start(ctx context.Context) error {
	p.manager.Logger().Info("Topaz plugin started: %s", p.config.Connection.Address)

	return nil
}

func (p *Plugin) Stop(ctx context.Context) {
	closeConn()

	p.manager.Logger().Info("Topaz plugin stopped")
}

func (p *Plugin) Reconfigure(ctx context.Context, c any) {
	if cfg, ok := c.(Config); ok {
		p.config = &cfg
		SetConfig(cfg)
	}
}

// func GetDirectoryConn() func() (*grpc.ClientConn, error) {
// 	return func() (*grpc.ClientConn, error) {
// 		return getConn()
// 	}
// }

func GetAccessClient() func() (access.AccessClient, error) {
	return func() (access.AccessClient, error) {
		conn, err := getConn()
		if err != nil {
			return nil, err
		}

		return access.NewAccessClient(conn), nil
	}
}

func GetDirectoryClient() func() (reader.ReaderClient, error) {
	return func() (reader.ReaderClient, error) {
		conn, err := getConn()
		if err != nil {
			return nil, err
		}

		return reader.NewReaderClient(conn), nil
	}
}
