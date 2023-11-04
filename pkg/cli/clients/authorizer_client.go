package clients

import (
	"context"

	azc "github.com/aserto-dev/go-aserto/client"
	"github.com/fullstorydev/grpcurl"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
)

type AuthorizerConfig struct {
	Host     string
	APIKey   string
	Insecure bool
	TenantID string
}

func NewAuthorizerClient(c *cc.CommonCtx, cfg *AuthorizerConfig) (authorizer.AuthorizerClient, error) {
	if cfg.Host == "" {
		cfg.Host = localhostAuthorizer
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	opts := []azc.ConnectionOption{
		azc.WithAddr(cfg.Host),
		azc.WithInsecure(cfg.Insecure),
	}

	if cfg.APIKey != "" {
		opts = append(opts, azc.WithAPIKeyAuth(cfg.APIKey))
	}

	if cfg.TenantID != "" {
		opts = append(opts, azc.WithTenantID(cfg.TenantID))
	}

	conn, err := azc.NewConnection(c.Context, opts...)
	if err != nil {
		return nil, err
	}

	return authorizer.NewAuthorizerClient(conn.Conn), nil
}

func (cfg *AuthorizerConfig) validate() error {
	ctx := context.Background()

	tlsConf, err := grpcurl.ClientTLSConfig(cfg.Insecure, "", "", "")
	if err != nil {
		return errors.Wrap(err, "failed to create TLS config")
	}

	creds := credentials.NewTLS(tlsConf)

	opts := []grpc.DialOption{
		grpc.WithUserAgent("topaz/dev-build (no version set)"),
	}
	if cfg.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	if _, err := grpcurl.BlockingDial(ctx, "tcp", cfg.Host, creds, opts...); err != nil {
		return err
	}
	return nil
}
