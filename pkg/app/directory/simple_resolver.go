package directory

import (
	"context"

	client "github.com/aserto-dev/go-aserto"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"google.golang.org/grpc"

	"github.com/aserto-dev/topaz/resolvers"
	"github.com/rs/zerolog"
)

type Resolver struct {
	logger  *zerolog.Logger
	cfg     *client.Config
	dirConn *grpc.ClientConn
}

var _ resolvers.DirectoryResolver = &Resolver{}

// The simple directory resolver returns a simple directory reader client.
func NewResolver(logger *zerolog.Logger, cfg *client.Config) resolvers.DirectoryResolver {
	return &Resolver{
		logger: logger,
		cfg:    cfg,
	}
}

func connect(logger *zerolog.Logger, cfg *client.Config) (*grpc.ClientConn, error) {
	logger.Debug().Str("tenant-id", cfg.TenantID).Str("addr", cfg.Address).Str("apiKey", cfg.APIKey).Bool("insecure", cfg.Insecure).Msg("GetDS")

	conn, err := client.NewConnection(
		client.WithAddr(cfg.Address),
		client.WithAPIKeyAuth(cfg.APIKey),
		client.WithTenantID(cfg.TenantID),
		client.WithInsecure(cfg.Insecure),
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// GetDS - returns a directory reader service client.
func (r *Resolver) GetDS(ctx context.Context) (dsr3.ReaderClient, error) {
	if r.dirConn == nil {
		dirConn, err := connect(r.logger, r.cfg)
		if err != nil {
			return nil, err
		}
		r.dirConn = dirConn
	}
	return dsr3.NewReaderClient(r.dirConn), nil
}
