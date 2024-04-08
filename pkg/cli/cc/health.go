package cc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// ServiceHealthStatus adopted from grpc-health-probe cli implementation
// https://github.com/grpc-ecosystem/grpc-health-probe/blob/master/main.go.
func ServiceHealthStatus(addr, service string) bool {
	rpcTimeout := time.Second * 30

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient(addr, dialOpts...)
	if err != nil {
		return false
	}
	defer conn.Close()

	rpcCtx, rpcCancel := context.WithTimeout(context.Background(), rpcTimeout)
	defer rpcCancel()

	if err := Retry(rpcTimeout, time.Millisecond*100, func() error {
		resp, err := healthpb.NewHealthClient(conn).Check(rpcCtx, &healthpb.HealthCheckRequest{Service: service})
		if err != nil {
			return err
		}

		if resp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
			return fmt.Errorf("gRPC endpoint not SERVING")
		}

		return nil
	}); err != nil {
		return false
	}

	return true
}
