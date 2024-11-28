package authorizer

import (
	"context"
	"time"

	client "github.com/aserto-dev/go-aserto"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	az2 "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
)

type Config struct {
	Host      string            `flag:"host" short:"H" default:"${authorizer_svc}" env:"TOPAZ_AUTHORIZER_SVC" help:"authorizer service address"`
	APIKey    string            `flag:"api-key" short:"k" default:"${authorizer_key}" env:"TOPAZ_AUTHORIZER_KEY" help:"authorizer API key"`
	Token     string            `flag:"token" default:"${authorizer_token}" env:"TOPAZ_AUTHORIZER_TOKEN" help:"authorizer OAuth2.0 token" hidden:""`
	Insecure  bool              `flag:"insecure" short:"i" default:"${insecure}" env:"TOPAZ_INSECURE" help:"skip TLS verification"`
	Plaintext bool              `flag:"plaintext" short:"P" default:"${plaintext}" env:"TOPAZ_PLAINTEXT" help:"use plain-text HTTP/2 (no TLS)"`
	TenantID  string            `flag:"tenant-id" help:"" default:"${tenant_id}" env:"ASERTO_TENANT_ID" `
	Headers   map[string]string `flag:"headers" env:"TOPAZ_AUTHORIZER_HEADERS" help:"additional headers to send to the authorizer service"`
	Timeout   time.Duration     `flag:"timeout" short:"T" default:"${timeout}" env:"TOPAZ_TIMEOUT" help:"command timeout"`
}

var _ clients.Config = &Config{}

type Client struct {
	conn       *grpc.ClientConn
	Authorizer az2.AuthorizerClient
}

func New(conn *grpc.ClientConn) *Client {
	return &Client{
		conn:       conn,
		Authorizer: az2.NewAuthorizerClient(conn),
	}
}

func NewClient(c *cc.CommonCtx, cfg *Config) (*Client, error) {
	conn, err := cfg.Connect(c.Context)
	if err != nil {
		return nil, err
	}

	return New(conn), nil
}

func (cfg *Config) Connect(ctx context.Context) (*grpc.ClientConn, error) {
	if cfg.Host == "" {
		return nil, errors.Errorf("no host specified")
	}

	if ok, err := clients.Validate(ctx, cfg); !ok {
		return nil, err
	}

	return cfg.ClientConfig().Connect()
}

func (cfg *Config) ClientConfig() *client.Config {
	return &client.Config{
		Address:  cfg.Host,
		Insecure: cfg.Insecure,
		NoTLS:    cfg.Plaintext,
		APIKey:   cfg.APIKey,
		Token:    cfg.Token,
		TenantID: cfg.TenantID,
		Headers:  cfg.Headers,
	}
}

func (cfg *Config) CommandTimeout() time.Duration {
	return cfg.Timeout
}
