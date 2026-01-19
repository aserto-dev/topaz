package cc

import (
	"context"
	"time"

	client "github.com/aserto-dev/go-aserto"
	"github.com/pkg/errors"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

const (
	rpcTimeout    = 10 * time.Second
	retryInterval = 100 * time.Millisecond
)

// ServiceHealthStatus adopted from grpc-health-probe cli implementation
// https://github.com/grpc-ecosystem/grpc-health-probe/blob/master/main.go.
func ServiceHealthStatus(ctx context.Context, cfg *client.Config, service string) (bool, error) {
	conn, err := cfg.Connect()
	if err != nil {
		return false, err
	}
	defer conn.Close()

	rpcCtx, rpcCancel := context.WithTimeout(ctx, rpcTimeout)
	defer rpcCancel()

	if err := Retry(rpcTimeout, retryInterval, func() error {
		resp, err := healthpb.NewHealthClient(conn).Check(rpcCtx, &healthpb.HealthCheckRequest{Service: service})
		if err != nil {
			return err
		}

		if resp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
			return errors.Errorf("health service %q is %s", service, resp.GetStatus().String())
		}

		return nil
	}); err != nil {
		return false, nil
	}

	return true, nil
}
