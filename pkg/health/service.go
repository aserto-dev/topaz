package health

import (
	"github.com/aserto-dev/topaz/pkg/servers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type Service struct {
	*health.Server
}

func New(cfg *Config) *Service {
	server := health.NewServer()

	for _, svc := range append(servers.KnownServices, servers.ServiceName("sync")) {
		server.SetServingStatus(string(svc), healthpb.HealthCheckResponse_NOT_SERVING)
	}

	return &Service{server}
}

func (s *Service) RegisterHealthServer(server *grpc.Server) {
	healthpb.RegisterHealthServer(server, s)
}
