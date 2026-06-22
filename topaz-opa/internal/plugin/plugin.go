package plugin

import (
	"context"

	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/topaz-opa/internal/config"
	"github.com/authzen/access.go/api/access/v1"
	"github.com/open-policy-agent/opa/v1/plugins"
	"google.golang.org/grpc"
)

type Plugin struct {
	manager *plugins.Manager
	config  *config.Config
}

var _ plugins.Plugin = (*Plugin)(nil)

func (p *Plugin) Start(ctx context.Context) error {
	p.manager.Logger().Info("Topaz plugin started: " + p.config.Connection.Address)

	return nil
}

func (p *Plugin) Stop(ctx context.Context) {
	p.manager.Logger().Info("Topaz plugin stopped")
}

func (p *Plugin) Reconfigure(ctx context.Context, c any) {
	if cfg, ok := c.(*config.Config); ok {
		p.config = cfg
	}
}

func GetDirectoryConn() func() (*grpc.ClientConn, error) {
	return func() (*grpc.ClientConn, error) {
		cfg := GetConfig()
		return cfg.Connection.Connect()
	}
}

func GetAccessClient() func() access.AccessClient {
	return func() access.AccessClient {
		cfg := GetConfig()

		conn, err := cfg.Connection.Connect()
		if err != nil {
			panic(err)
		}

		return access.NewAccessClient(conn)
	}
}

func GetDirectoryClient() func() reader.ReaderClient {
	return func() reader.ReaderClient {
		cfg := GetConfig()

		conn, err := cfg.Connection.Connect()
		if err != nil {
			panic(err)
		}

		return reader.NewReaderClient(conn)
	}
}
