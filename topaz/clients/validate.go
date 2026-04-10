package clients

import (
	"context"
	"io"
	"time"

	client "github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/topaz/topaz/version"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection/grpc_reflection_v1"
	"google.golang.org/grpc/status"
)

const defaultValidationTimeout = time.Second * 10

func Validate(ctx context.Context, cfg Config) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultValidationTimeout)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	clientCfg := cfg.ClientConfig()

	conn, err := clientCfg.Connect(
		client.WithDialOptions(
			grpc.WithUserAgent(version.UserAgent()),
		),
	)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	refClient := grpc_reflection_v1.NewServerReflectionClient(conn)

	stream, err := refClient.ServerReflectionInfo(ctx)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			return false, errors.Errorf("%s: %s", st.Code().String(), st.Message())
		}

		return false, err
	}

	// Receiver
	g.Go(func() error {
		for {
			in, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}

				return err
			}

			if in.GetValidHost() == "" {
				return status.Errorf(codes.Unavailable, "no valid host")
			}

			// Only need first valid response
			return nil //nolint:staticcheck
		}
	})

	// Sender
	if err := stream.Send(&grpc_reflection_v1.ServerReflectionRequest{
		Host:           clientCfg.Address,
		MessageRequest: &grpc_reflection_v1.ServerReflectionRequest_ListServices{},
	}); err != nil {
		cancel()
		return false, err
	}

	if err := stream.CloseSend(); err != nil {
		return false, err
	}

	if err := g.Wait(); err != nil {
		return false, err
	}

	return true, nil
}
