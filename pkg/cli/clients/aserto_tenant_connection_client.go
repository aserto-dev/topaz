package clients

import (
	"errors"

	grpcClient "github.com/aserto-dev/go-aserto/client"
	"github.com/aserto-dev/go-grpc/aserto/tenant/connection/v1"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
)

type TenantConfig struct {
	Address  string
	APIKey   string
	Token    string
	TenantID string
	Insecure bool
}

func NewTenantConnectionClient(c *cc.CommonCtx, cfg *TenantConfig) (connection.ConnectionClient, error) {
	if cfg.Address == "" {
		return nil, errors.New("tenant address not specified")
	}

	opts := []grpcClient.ConnectionOption{
		grpcClient.WithAddr(cfg.Address),
		grpcClient.WithInsecure(cfg.Insecure),
	}

	if cfg.Token != "" {
		opts = append(opts, grpcClient.WithTokenAuth(cfg.Token))
	}
	if cfg.APIKey != "" {
		opts = append(opts, grpcClient.WithAPIKeyAuth(cfg.APIKey))
	}

	if cfg.TenantID != "" {
		opts = append(opts, grpcClient.WithTenantID(cfg.TenantID))
	}

	conn, err := grpcClient.NewConnection(c.Context, opts...)
	if err != nil {
		return nil, err
	}

	return connection.NewConnectionClient(conn.Conn), nil
}
