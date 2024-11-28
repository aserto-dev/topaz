package directory

import (
	"context"
	"time"

	client "github.com/aserto-dev/go-aserto"
	dsa3 "github.com/aserto-dev/go-directory/aserto/directory/assertion/v3"
	dse3 "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	dsi3 "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	dsm3 "github.com/aserto-dev/go-directory/aserto/directory/model/v3"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	dsw3 "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type Config struct {
	Host      string            `flag:"host" short:"H" default:"${directory_svc}" env:"TOPAZ_DIRECTORY_SVC" help:"directory service address"`
	APIKey    string            `flag:"api-key" short:"k" default:"${directory_key}" env:"TOPAZ_DIRECTORY_KEY" help:"directory API key"`
	Token     string            `flag:"token" default:"${directory_token}" env:"TOPAZ_DIRECTORY_TOKEN" help:"directory OAuth2.0 token" hidden:""`
	Insecure  bool              `flag:"insecure" short:"i" default:"${insecure}" env:"TOPAZ_INSECURE" help:"skip TLS verification"`
	Plaintext bool              `flag:"plaintext" short:"P" default:"${plaintext}" env:"TOPAZ_PLAINTEXT" help:"use plain-text HTTP/2 (no TLS)"`
	TenantID  string            `flag:"tenant-id" help:"" default:"${tenant_id}" env:"ASERTO_TENANT_ID" `
	Headers   map[string]string `flag:"headers" env:"TOPAZ_DIRECTORY_HEADERS" help:"additional headers to send to the directory service"`
	Timeout   time.Duration     `flag:"timeout" short:"T" default:"${timeout}" env:"TOPAZ_TIMEOUT" help:"command timeout"`
}

var _ clients.Config = &Config{}

type Client struct {
	conn      *grpc.ClientConn
	Model     dsm3.ModelClient
	Reader    dsr3.ReaderClient
	Writer    dsw3.WriterClient
	Importer  dsi3.ImporterClient
	Exporter  dse3.ExporterClient
	Assertion dsa3.AssertionClient
}

func New(conn *grpc.ClientConn) *Client {
	return &Client{
		conn:      conn,
		Model:     dsm3.NewModelClient(conn),
		Reader:    dsr3.NewReaderClient(conn),
		Writer:    dsw3.NewWriterClient(conn),
		Importer:  dsi3.NewImporterClient(conn),
		Exporter:  dse3.NewExporterClient(conn),
		Assertion: dsa3.NewAssertionClient(conn),
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
