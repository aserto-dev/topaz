package directory

import (
	"context"
	"fmt"

	"github.com/aserto-dev/go-aserto/client"
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
	Host     string `flag:"host" short:"H" default:"${directory_svc}" env:"TOPAZ_DIRECTORY_SVC" help:"directory service address"`
	APIKey   string `flag:"api-key" short:"k" default:"${directory_key}" env:"TOPAZ_DIRECTORY_KEY" help:"directory API key"`
	Token    string `flag:"token" default:"${directory_token}" env:"TOPAZ_DIRECTORY_TOKEN" help:"directory OAuth2.0 token" hidden:""`
	Insecure bool   `flag:"insecure" short:"i" default:"${insecure}" env:"TOPAZ_INSECURE" help:"skip TLS verification"`
	TenantID string `flag:"tenant-id" help:"" default:"${tenant_id}" env:"ASERTO_TENANT_ID" `
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

func NewConn(ctx context.Context, cfg *Config) (*grpc.ClientConn, error) {
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
	conn, err := NewConn(c.Context, cfg)
	if err != nil {
		return nil, err
	}

	return New(conn), nil
}

func (cfg *Config) validate() error {
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
