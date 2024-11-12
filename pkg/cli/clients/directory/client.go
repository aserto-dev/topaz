package directory

import (
	"context"

	client "github.com/aserto-dev/go-aserto"
	dsa3 "github.com/aserto-dev/go-directory/aserto/directory/assertion/v3"
	dse3 "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	dsi3 "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	dsm3 "github.com/aserto-dev/go-directory/aserto/directory/model/v3"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	dsw3 "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	"github.com/aserto-dev/topaz/pkg/cli/cc"

	"github.com/fullstorydev/grpcurl"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	Host     string            `flag:"host" short:"H" default:"${directory_svc}" env:"TOPAZ_DIRECTORY_SVC" help:"directory service address"`
	APIKey   string            `flag:"api-key" short:"k" default:"${directory_key}" env:"TOPAZ_DIRECTORY_KEY" help:"directory API key"`
	Token    string            `flag:"token" default:"${directory_token}" env:"TOPAZ_DIRECTORY_TOKEN" help:"directory OAuth2.0 token" hidden:""`
	Insecure bool              `flag:"insecure" short:"i" default:"${insecure}" env:"TOPAZ_INSECURE" help:"skip TLS verification"`
	TenantID string            `flag:"tenant-id" help:"" default:"${tenant_id}" env:"ASERTO_TENANT_ID" `
	Headers  map[string]string `flag:"headers" env:"TOPAZ_DIRECTORY_HEADERS" help:"additional headers to send to the directory service"`
}

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

	if err := cfg.validate(ctx); err != nil {
		return nil, err
	}

	return cfg.connect()
}

func (cfg *Config) validate(ctx context.Context) error {
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

func (cfg *Config) connect() (*grpc.ClientConn, error) {
	clientCfg := &client.Config{
		Address:  cfg.Host,
		Insecure: cfg.Insecure,
		APIKey:   cfg.APIKey,
		Token:    cfg.Token,
		TenantID: cfg.TenantID,
		Headers:  cfg.Headers,
	}

	return clientCfg.Connect()
}
