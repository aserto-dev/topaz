package directory

import (
	"context"

	grpcc "github.com/aserto-dev/go-aserto/client"
	ds2 "github.com/aserto-dev/go-directory/aserto/directory/reader/v2"

	"github.com/aserto-dev/topaz/directory"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/rs/zerolog"
)

type Resolver struct {
	logger  *zerolog.Logger
	cfg     *directory.Config
	dirConn *grpcc.Connection
}

var _ resolvers.DirectoryResolver = &Resolver{}

func NewResolver(logger *zerolog.Logger, cfg *directory.Config) resolvers.DirectoryResolver {
	return &Resolver{
		logger: logger,
		cfg:    cfg,
	}
}

func connect(logger *zerolog.Logger, cfg *directory.Config) (*grpcc.Connection, error) {
	logger.Debug().Str("tenant-id", cfg.Remote.TenantID).Str("addr", cfg.Remote.Addr).Str("apiKey", cfg.Remote.Key).Bool("insecure", cfg.Remote.Insecure).Msg("GetDS")

	ctx := context.Background()

	conn, err := grpcc.NewConnection(ctx,
		grpcc.WithAddr(cfg.Remote.Addr),
		grpcc.WithAPIKeyAuth(cfg.Remote.Key),
		grpcc.WithTenantID(cfg.Remote.TenantID),
		grpcc.WithInsecure(cfg.Remote.Insecure),
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// GetDS - simple
//
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
