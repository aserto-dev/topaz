package clients

import (
	"context"

	grpcClient "github.com/aserto-dev/go-aserto/client"
	"github.com/aserto-dev/go-directory-cli/client"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/fullstorydev/grpcurl"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const localhostDirectory = "localhost:9292"

func NewDirectoryClient(c *cc.CommonCtx, cfg *Config) (*client.Client, error) {

	if cfg.Host == "" {
		cfg.Host = localhostDirectory
	}

	if err := validate(cfg); err != nil {
		return nil, err
	}

	opts := []grpcClient.ConnectionOption{
		grpcClient.WithAddr(cfg.Host),
		grpcClient.WithInsecure(cfg.Insecure),
	}

	if cfg.APIKey != "" {
		opts = append(opts, grpcClient.WithAPIKeyAuth(cfg.APIKey))
	}

	if cfg.SessionID != "" {
		opts = append(opts, grpcClient.WithSessionID(cfg.SessionID))
	}

	if cfg.TenantID != "" {
		opts = append(opts, grpcClient.WithTenantID(cfg.TenantID))
	}

	conn, err := grpcClient.NewConnection(c.Context, opts...)
	if err != nil {
		return nil, err
	}

	return client.New(conn.Conn, c.UI)
}

type Config struct {
	Host      string `flag:"host" short:"H" help:"" env:"TOPAZ_DIRECTORY_SVC" default:"localhost:9292"`
	APIKey    string `flag:"api-key" short:"k" help:"" env:"TOPAZ_DIRECTORY_KEY"`
	Insecure  bool   `flag:"insecure" short:"i" help:""`
	SessionID string `flag:"session-id"  help:""`
	TenantID  string `flag:"tenant-id" help:""`
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
