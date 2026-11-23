package client

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// ConnectionOptions holds settings used to establish a connection to the authorizer service.
type ConnectionOptions struct {
	Config

	// Credentials used to authenticate with the authorizer service. Either API Key or OAuth Token.
	Creds credentials.PerRPCCredentials

	// UnaryClientInterceptors passed to the grpc client.
	UnaryClientInterceptors []grpc.UnaryClientInterceptor

	// StreamClientInterceptors passed to the grpc client.
	StreamClientInterceptors []grpc.StreamClientInterceptor

	// DialOptions passed to the grpc client.
	DialOptions []grpc.DialOption
}

// NewConnectionOptions creates a ConnectionOptions object from a collection of ConnectionOption functions.
func NewConnectionOptions(opts ...ConnectionOption) (*ConnectionOptions, error) {
	options := &ConnectionOptions{}
	if err := options.Apply(opts...); err != nil {
		return nil, err
	}

	return options, nil
}

// Apply additional options.
func (o *ConnectionOptions) Apply(opts ...ConnectionOption) error {
	var errs error

	for _, opt := range opts {
		if err := opt(o); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs
}

func (o *ConnectionOptions) ToDialOptions() ([]grpc.DialOption, error) {
	if o.Insecure && o.NoTLS {
		return nil, errors.Wrap(ErrInvalidOptions, "insecure and no_tls options are mutually exclusive")
	}

	transportCreds, err := o.transportCredentials()
	if err != nil {
		return nil, err
	}

	opts := []grpc.DialOption{
		transportCreds,
		grpc.WithChainStreamInterceptor(o.StreamClientInterceptors...),
		grpc.WithChainUnaryInterceptor(o.UnaryClientInterceptors...),
	}

	opts = append(opts, o.DialOptions...)

	if o.Creds != nil {
		opts = append(opts, grpc.WithPerRPCCredentials(o.Creds))
	}

	if o.NoProxy {
		opts = append(opts, grpc.WithNoProxy())
	}

	if len(o.Headers) > 0 {
		opts = append(opts, o.outgoingHeaders()...)
	}

	return opts, nil
}

func (o *ConnectionOptions) transportCredentials() (grpc.DialOption, error) {
	if o.NoTLS {
		return grpc.WithTransportCredentials(insecure.NewCredentials()), nil
	}

	cfg := &TLSConfig{
		Cert: o.ClientCertPath,
		Key:  o.ClientKeyPath,
		CA:   o.CACertPath,
	}

	creds, err := cfg.ClientCredentials(o.Insecure)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create transport credentials")
	}

	return grpc.WithTransportCredentials(creds), nil
}

func (o *ConnectionOptions) outgoingHeaders() []grpc.DialOption {
	pairs := lo.Reduce(
		lo.Entries(o.Headers),
		func(acc []string, entry lo.Entry[string, string], _ int) []string {
			return append(acc, entry.Key, entry.Value)
		},
		nil,
	)

	appendOutgoing := func(ctx context.Context) context.Context {
		return metadata.AppendToOutgoingContext(ctx, pairs...)
	}

	return contextWrapperInterceptor(appendOutgoing)
}

func contextWrapperInterceptor(wrap func(ctx context.Context) context.Context) []grpc.DialOption {
	unary := func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		return invoker(wrap(ctx), method, req, reply, cc, opts...)
	}

	stream := func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		return streamer(wrap(ctx), desc, cc, method, opts...)
	}

	return []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(unary),
		grpc.WithChainStreamInterceptor(stream),
	}
}
