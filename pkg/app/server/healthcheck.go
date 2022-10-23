package server

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// HealthServer contains everything we need to be able to serve a health status endpoint
type HealthServer struct {
	Server     *health.Server
	GRPCServer *grpc.Server
}

// newGRPCHealthServer creates a new HealthServer.
func newGRPCHealthServer() *HealthServer {
	healthServer := health.NewServer()
	grpcHealthServer := grpc.NewServer()

	healthpb.RegisterHealthServer(grpcHealthServer, healthServer)

	return &HealthServer{
		Server:     healthServer,
		GRPCServer: grpcHealthServer,
	}
}
