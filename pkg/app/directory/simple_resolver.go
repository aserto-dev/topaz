package directory

import (
	"context"

	grpcc "github.com/aserto-dev/aserto-go/client"
	ds2 "github.com/aserto-dev/go-directory/aserto/directory/v2"

	eds "github.com/aserto-dev/go-eds"
	"github.com/aserto-dev/go-lib/ids"
	"github.com/aserto-dev/topaz/directory"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type Resolver struct {
	logger    *zerolog.Logger
	directory *eds.EdgeDirectory
	dirConn   *grpcc.Connection
}

var _ resolvers.DirectoryResolver = &Resolver{}

func NewResolver(logger *zerolog.Logger, cfg *directory.Config) (resolvers.DirectoryResolver, func(), error) {
	dir, cleanup, err := eds.NewEdgeDirectory(cfg.EDSPath(), logger)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create directory resolver")
	}

	err = dir.Open()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to open directory")
	}

	dirConn, err := connect(logger, cfg)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed create directory client")
	}

	return &Resolver{
		logger:    logger,
		directory: dir,
		dirConn:   dirConn,
	}, cleanup, nil
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
func (r *Resolver) GetDS(ctx context.Context) (ds2.DirectoryClient, error) {
	return ds2.NewDirectoryClient(r.dirConn.Conn), nil
}

func (r *Resolver) DirectoryFromContext(ctx context.Context) (directory.Directory, error) {
	tenantID := ids.ExtractTenantID(ctx)
	return r.GetDirectory(ctx, tenantID)
}

func (r *Resolver) GetDirectory(ctx context.Context, instanceID string) (directory.Directory, error) {
	return r.directory, nil
}

func (r *Resolver) ReloadDirectory(ctx context.Context, instanceID string) error {
	return nil
}

func (r *Resolver) ListDirectories(ctx context.Context) ([]string, error) {
	return r.directory.ListTenants()
}

func (r *Resolver) RemoveDirectory(ctx context.Context, instanceID string) error {
	return nil
}
