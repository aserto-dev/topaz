package clients

import (
	grpcClient "github.com/aserto-dev/aserto-go/client"
	"github.com/aserto-dev/go-directory-cli/client"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
)

const localhostDirectory = "localhost:9292"

func NewDirectoryClient(c *cc.CommonCtx, cfg *Config) (*client.Client, error) {

	if cfg.Host == "" {
		cfg.Host = localhostDirectory
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

	return client.New(conn, c.UI)
}

type Config struct {
	Host      string `flag:"host" short:"H" help:"" env:"TOPAZ_DIRECTORY_SVC" default:"localhost:9292"`
	APIKey    string `flag:"api-key" short:"k" help:"" env:"TOPAZ_DIRECTORY_KEY"`
	Insecure  bool   `flag:"insecure" short:"i" help:""`
	SessionID string `flag:"session-id"  help:""`
	TenantID  string `flag:"tenant-id" help:""`
}
