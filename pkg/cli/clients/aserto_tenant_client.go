package clients

import (
	"errors"

	"github.com/aserto-dev/aserto-go/client"
	"github.com/aserto-dev/aserto-go/client/tenant"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
)

type TenantConfig struct {
	Address  string
	APIKey   string
	TenantID string
	Insecure bool
}

func NewTenantClient(c *cc.CommonCtx, cfg *TenantConfig) (*tenant.Client, error) {
	if cfg.Address == "" {
		return nil, errors.New("tenant address not specified")
	}

	opts := []client.ConnectionOption{
		client.WithAddr(cfg.Address),
		client.WithInsecure(cfg.Insecure),
	}

	if cfg.APIKey != "" {
		opts = append(opts, client.WithAPIKeyAuth(cfg.APIKey))
	}

	if cfg.TenantID != "" {
		opts = append(opts, client.WithTenantID(cfg.TenantID))
	}

	return tenant.New(c.Context, opts...)
}
