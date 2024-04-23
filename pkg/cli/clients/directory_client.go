package clients

import (
	"context"

	"github.com/aserto-dev/go-aserto/client"
	dsc "github.com/aserto-dev/go-directory-cli/client"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/fullstorydev/grpcurl"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	localhostDirectory   string = "localhost:9292"
	EnvTopazDirectorySvc string = "TOPAZ_DIRECTORY_SVC"
	EnvTopazDirectoryKey string = "TOPAZ_DIRECTORY_KEY"
)

type DirectoryConfig struct {
	Host     string `flag:"host" short:"H" env:"TOPAZ_DIRECTORY_SVC" help:"directory service address"`
	APIKey   string `flag:"api-key" short:"k" env:"TOPAZ_DIRECTORY_KEY" help:"directory API key"`
	Token    string `flag:"token" short:"t" env:"TOPAZ_DIRECTORY_TOKEN" help:"directory OAuth2.0 token" hidden:""`
	Insecure bool   `flag:"insecure" short:"i" env:"INSECURE" help:"skip TLS verification"`
	TenantID string `flag:"tenant-id" help:"" env:"ASERTO_TENANT_ID" `
}

func NewDirectoryClient(c *cc.CommonCtx, cfg *DirectoryConfig) (*dsc.Client, error) {

	if cfg.Host == "" {
		cfg.Host = localhostDirectory
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

	conn, err := client.NewConnection(c.Context, opts...)
	if err != nil {
		return nil, err
	}

	return dsc.New(conn, c.UI)
}

func (cfg *DirectoryConfig) validate() error {
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
