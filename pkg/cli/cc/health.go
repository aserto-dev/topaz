package cc

import (
	"context"
	"time"

	client "github.com/aserto-dev/go-aserto"
	"github.com/pkg/errors"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

const rpcTimeout = time.Second * 30

// ServiceHealthStatus adopted from grpc-health-probe cli implementation
// https://github.com/grpc-ecosystem/grpc-health-probe/blob/master/main.go.
func ServiceHealthStatus(ctx context.Context, addr, service string) (bool, error) {
	conn, err := client.NewConnection(
		client.WithAddr(addr),
		client.WithInsecure(true),
	)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	rpcCtx, rpcCancel := context.WithTimeout(ctx, rpcTimeout)
	defer rpcCancel()

	if err := Retry(rpcTimeout, time.Millisecond*100, func() error {
		resp, err := healthpb.NewHealthClient(conn).Check(rpcCtx, &healthpb.HealthCheckRequest{Service: service})
		if err != nil {
			return err
		}

		if resp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
			return errors.Errorf("gRPC endpoint not SERVING")
		}

		return nil
	}); err != nil {
		return false, nil
	}

	return true, nil
}
