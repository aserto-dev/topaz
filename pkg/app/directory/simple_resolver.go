package directory

import (
	"context"

	grpcc "github.com/aserto-dev/go-aserto/client"
	ds2 "github.com/aserto-dev/go-directory/aserto/directory/reader/v2"

	"github.com/aserto-dev/topaz/resolvers"
	"github.com/rs/zerolog"
)

type Resolver struct {
	logger  *zerolog.Logger
	cfg     *grpcc.Config
	dirConn *grpcc.Connection
}

var _ resolvers.DirectoryResolver = &Resolver{}

// The simple directory resolver returns a simple directory reader client.
func NewResolver(logger *zerolog.Logger, cfg *grpcc.Config) resolvers.DirectoryResolver {
	return &Resolver{
		logger: logger,
		cfg:    cfg,
	}
}

func connect(logger *zerolog.Logger, cfg *grpcc.Config) (*grpcc.Connection, error) {
	logger.Debug().Str("tenant-id", cfg.TenantID).Str("addr", cfg.Address).Str("apiKey", cfg.APIKey).Bool("insecure", cfg.Insecure).Msg("GetDS")

	ctx := context.Background()

	conn, err := grpcc.NewConnection(ctx,
		grpcc.WithAddr(cfg.Address),
		grpcc.WithAPIKeyAuth(cfg.APIKey),
		grpcc.WithTenantID(cfg.TenantID),
		grpcc.WithInsecure(cfg.Insecure),
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// GetDS - returns a directory reader service client.
func (r *Resolver) GetDS(ctx context.Context) (ds2.ReaderClient, error) {
	if r.dirConn == nil {
		dirConn, err := connect(r.logger, r.cfg)
		if err != nil {
			return nil, err
		}
		r.dirConn = dirConn
	}
	return ds2.NewReaderClient(r.dirConn.Conn), nil
}
