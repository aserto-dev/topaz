package builder

import (
	"github.com/aserto-dev/go-aserto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Health struct {
	Server     *health.Server
	GRPCServer *grpc.Server
	Address    string
}

// newGRPCHealthServer creates a new HealthServer.
func newGRPCHealthServer(certCfg *aserto.TLSConfig) *Health {
	healthServer := health.NewServer()

	grpcHealthServer, err := prepareGrpcServer(certCfg, nil)
	if err != nil {
		panic(err)
	}

	healthpb.RegisterHealthServer(grpcHealthServer, healthServer)
	reflection.Register(grpcHealthServer)
	return &Health{
		Server:     healthServer,
		GRPCServer: grpcHealthServer,
	}
}

func (h *Health) SetServiceStatus(service string, status healthpb.HealthCheckResponse_ServingStatus) {
	h.Server.SetServingStatus(service, status)
}
