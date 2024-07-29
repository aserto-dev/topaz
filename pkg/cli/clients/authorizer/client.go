package authorizer

import (
	"context"
	"fmt"

	"github.com/aserto-dev/go-aserto/client"
	"github.com/fullstorydev/grpcurl"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	az2 "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
)

type AuthorizerConfig struct {
	Host     string `flag:"host" short:"H" default:"${authorizer_svc}" env:"TOPAZ_AUTHORIZER_SVC" help:"authorizer service address"`
	APIKey   string `flag:"api-key" short:"k" default:"${authorizer_key}" env:"TOPAZ_AUTHORIZER_KEY" help:"authorizer API key"`
	Token    string `flag:"token" default:"${authorizer_token}" env:"TOPAZ_AUTHORIZER_TOKEN" help:"authorizer OAuth2.0 token" hidden:""`
	Insecure bool   `flag:"insecure" short:"i" default:"${insecure}" env:"TOPAZ_INSECURE" help:"skip TLS verification"`
	TenantID string `flag:"tenant-id" help:"" default:"${tenant_id}" env:"ASERTO_TENANT_ID" `
}

type Client struct {
	conn       *grpc.ClientConn
	Authorizer az2.AuthorizerClient
}

func NewConn(ctx context.Context, cfg *AuthorizerConfig) (*grpc.ClientConn, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("no host specified")
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	opts := []client.ConnectionOption{
		client.WithAddr(cfg.Host),
		client.WithInsecure(cfg.Insecure),
	}

	if cfg.APIKey != "" {
		opts = append(opts, client.WithAPIKeyAuth(cfg.APIKey))
	}

	if cfg.Token != "" {
		opts = append(opts, client.WithTokenAuth(cfg.Token))
	}

	if cfg.TenantID != "" {
		opts = append(opts, client.WithTenantID(cfg.TenantID))
	}

	return client.NewConnection(ctx, opts...)
}

func New(conn *grpc.ClientConn) *Client {
	return &Client{
		conn:       conn,
		Authorizer: az2.NewAuthorizerClient(conn),
	}
}

func NewClient(c *cc.CommonCtx, cfg *AuthorizerConfig) (*Client, error) {
	conn, err := NewConn(c.Context, cfg)
	if err != nil {
		return nil, err
	}

	return New(conn), nil
}

func (cfg *AuthorizerConfig) validate() error {
	ctx := context.Background()

	tlsConf, err := grpcurl.ClientTLSConfig(cfg.Insecure, "", "", "")
	if err != nil {
		return errors.Wrap(err, "failed to create TLS config")
	}

	creds := credentials.NewTLS(tlsConf)

	opts := []grpc.DialOption{
		grpc.WithUserAgent("topaz/dev-build (no version set)"), // TODO: verify user-agent value
	}

	if cfg.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if _, err := grpcurl.BlockingDial(ctx, "tcp", cfg.Host, creds, opts...); err != nil {
		return err
	}

	return nil
}
