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

type Config struct {
	Host      string `flag:"host" short:"H" help:"directory service address" env:"TOPAZ_DIRECTORY_SVC" default:"localhost:9292"`
	APIKey    string `flag:"api-key" short:"k" help:"directory API key" env:"TOPAZ_DIRECTORY_KEY"`
	Token     string `flag:"token" short:"t" help:"JWT used for connection" env:"TOPAZ_DIRECTORY_TOKEN"`
	Insecure  bool   `flag:"insecure" short:"i" help:"skip TLS verification" env:"TOPAZ_DIRECTORY_INSECURE"`
	SessionID string `flag:"session-id"  help:"" env:"TOPAZ_SESSION_ID"`
	TenantID  string `flag:"tenant-id" help:"" env:"TOPAZ_TENANT_ID"`
}

func NewDirectoryClient(c *cc.CommonCtx, cfg *Config) (*dsc.Client, error) {

	if cfg.Host == "" {
		cfg.Host = localhostDirectory
	}

	if err := validate(cfg); err != nil {
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

	if cfg.SessionID != "" {
		opts = append(opts, client.WithSessionID(cfg.SessionID))
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

func validate(cfg *Config) error {
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
