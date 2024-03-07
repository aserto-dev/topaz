package clients

import (
	azc "github.com/aserto-dev/go-aserto/client"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
)

func NewAuthorizerClient(c *cc.CommonCtx, cfg *Config) (authorizer.AuthorizerClient, error) {
	if cfg.Host == "" {
		cfg.Host = localhostAuthorizer
	}

	if err := validate(cfg); err != nil {
		return nil, err
	}

	opts := []azc.ConnectionOption{
		azc.WithAddr(cfg.Host),
		azc.WithInsecure(cfg.Insecure),
	}

	if cfg.APIKey != "" {
		opts = append(opts, azc.WithAPIKeyAuth(cfg.APIKey))
	}

	if cfg.Token != "" {
		opts = append(opts, azc.WithTokenAuth(cfg.Token))
	}

	if cfg.TenantID != "" {
		opts = append(opts, azc.WithTenantID(cfg.TenantID))
	}

	conn, err := azc.NewConnection(c.Context, opts...)
	if err != nil {
		return nil, err
	}

	return authorizer.NewAuthorizerClient(conn), nil
}
