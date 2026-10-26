package topaz

import (
	"context"

	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/topaz-opa/internal/errs"
	"github.com/authzen/access.go/api/access/v1"
	"github.com/open-policy-agent/opa/v1/plugins"

	"google.golang.org/grpc"
)

const PluginName string = `topaz`

type Plugin struct {
	manager *plugins.Manager
	config  *Config
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
	if cfg, ok := c.(*Config); ok {
		p.config = cfg
	}
}

func GetDirectoryConn() func() (*grpc.ClientConn, error) {
	return func() (*grpc.ClientConn, error) {
		cfg := GetConfig()
		if !cfg.Enabled {
			return nil, errs.ErrTopazPluginDisabled
		}

		return cfg.Connection.Connect()
	}
}

func GetAccessClient() func() (access.AccessClient, error) {
	return func() (access.AccessClient, error) {
		cfg := GetConfig()
		if !cfg.Enabled {
			return nil, errs.ErrTopazPluginDisabled
		}

		conn, err := cfg.Connection.Connect()
		if err != nil {
			return nil, err
		}

		return access.NewAccessClient(conn), nil
	}
}

func GetDirectoryClient() func() (reader.ReaderClient, error) {
	return func() (reader.ReaderClient, error) {
		cfg := GetConfig()
		if !cfg.Enabled {
			return nil, errs.ErrTopazPluginDisabled
		}

		conn, err := cfg.Connection.Connect()
		if err != nil {
			return nil, err
		}

		return reader.NewReaderClient(conn), nil
	}
}
