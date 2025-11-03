package directory

import (
	client "github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/authzen/access.go/api/access/v1"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

type Resolver struct {
	dirConn *grpc.ClientConn
	logger  *zerolog.Logger
}

var _ resolvers.DirectoryResolver = &Resolver{}

// NewResolver returns a simple directory reader client.
func NewResolver(logger *zerolog.Logger, cfg *client.Config) (*Resolver, error) {
	l := logger.With().Interface("client", cfg).Logger()
	l.Debug().Msg("new directory resolver")

	conn, err := cfg.Connect()
	if err != nil {
		return nil, err
	}

	return &Resolver{dirConn: conn, logger: &l}, nil
}

func (r *Resolver) Close() {
	if err := r.dirConn.Close(); err != nil {
		r.logger.Err(err).Msg("failed to close directory connection")
	}
}

// GetDS - returns a directory reader service client.
func (r *Resolver) GetDS() reader.ReaderClient {
	return reader.NewReaderClient(r.dirConn)
}

// GetAuthZen - returns an AuthZen access service client.
func (r *Resolver) GetAuthZen() access.AccessClient {
	return access.NewAccessClient(r.dirConn)
}
