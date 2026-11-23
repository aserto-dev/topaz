package client

import (
	"github.com/aserto-dev/topaz/api/client/internal/hosted"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// NewConnection creates a gRPC connection with the given options.
func NewConnection(opts ...ConnectionOption) (*grpc.ClientConn, error) {
	options, err := NewConnectionOptions(opts...)
	if err != nil {
		return nil, err
	}

	if options.Address == "" {
		// Backward compatibility: default to authorizer service.
		options.Address = hosted.HostedAuthorizerHostname + hosted.HostedAuthorizerGRPCPort
	}

	return Connect(options)
}

// Connect creates a gRPC connection with the given options.
func Connect(options *ConnectionOptions) (*grpc.ClientConn, error) {
	if options.Address == "" {
		return nil, errors.Wrap(ErrInvalidOptions, "address not specified")
	}

	dialOpts, err := options.ToDialOptions()
	if err != nil {
		return nil, err
	}

	return grpc.NewClient(options.Address, dialOpts...)
}
