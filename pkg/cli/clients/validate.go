package clients

import (
	"context"
	"io"

	client "github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/topaz/pkg/version"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection/grpc_reflection_v1"
	"google.golang.org/grpc/status"
)

func Validate(ctx context.Context, cfg Config) (bool, error) {
	g, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithTimeout(ctx, cfg.CommandTimeout())
	defer cancel()

	clientCfg := cfg.ClientConfig()

	conn, err := clientCfg.Connect(
		client.WithDialOptions(
			grpc.WithUserAgent(version.UserAgent()),
		),
	)
	if err != nil {
		return false, err
	}

	callOpts := []grpc.CallOption{}

	refClient := grpc_reflection_v1.NewServerReflectionClient(conn)
	stream, err := refClient.ServerReflectionInfo(ctx, callOpts...)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			return false, errors.Errorf("%s: %s", st.Code().String(), st.Message())
		}
		return false, err
	}

	g.Go(func() error {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		if in.GetValidHost() == "" {
			return status.Errorf(codes.Unavailable, "no valid host")
		}
		return nil
	})

	if err := stream.Send(&grpc_reflection_v1.ServerReflectionRequest{
		Host:           clientCfg.Address,
		MessageRequest: &grpc_reflection_v1.ServerReflectionRequest_ListServices{},
	}); err != nil {
		return false, err
	}

	_ = stream.CloseSend()

	if err := g.Wait(); err != nil {
		return false, err
	}

	return true, nil
}
