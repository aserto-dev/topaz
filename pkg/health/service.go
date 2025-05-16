package health

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type Service health.Server

func New(cfg *Config) *Service {
	return (*Service)(health.NewServer())
}

func (s *Service) RegisterHealthServer(server *grpc.Server) {
	healthpb.RegisterHealthServer(server, s)
}
