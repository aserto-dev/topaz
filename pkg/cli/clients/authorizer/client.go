package authorizer

import (
	"context"
	"time"

	client "github.com/aserto-dev/go-aserto"
	"github.com/fullstorydev/grpcurl"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	az2 "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/version"
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

	if err := cfg.validate(ctx); err != nil {
		return nil, err
	}

	return cfg.ClientConfig().Connect()
}

func (cfg *Config) validate(ctx context.Context) error {
	var creds credentials.TransportCredentials
	if cfg.Insecure {
		tlsConf, err := grpcurl.ClientTLSConfig(cfg.Insecure, "", "", "")
		if err != nil {
			return errors.Wrap(err, "failed to create TLS config")
		}
		creds = credentials.NewTLS(tlsConf)
	}

	opts := []grpc.DialOption{
		grpc.WithUserAgent(version.UserAgent()),
	}

	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if _, err := grpcurl.BlockingDial(ctx, "tcp", cfg.Host, creds, opts...); err != nil {
		return err
	}

	return nil
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
