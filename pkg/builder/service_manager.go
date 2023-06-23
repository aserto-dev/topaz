package builder

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type ServiceManager struct {
	Context  context.Context
	logger   *zerolog.Logger
	errGroup *errgroup.Group

	GRPCServers map[string]*Server
}

type HandlerRegistrations func(ctx context.Context, mux *runtime.ServeMux, grpcEndpoint string, opts []grpc.DialOption) error

func NewServiceManager(logger *zerolog.Logger) *ServiceManager {
	errGroup, ctx := errgroup.WithContext(context.Background())
	return &ServiceManager{
		Context:     ctx,
		logger:      logger,
		GRPCServers: make(map[string]*Server),
		errGroup:    errGroup,
	}
}

func (s *ServiceManager) AddGRPCServer(server *Server) error {
	s.GRPCServers[server.Config.GRPC.ListenAddress] = server
	return nil
}

func (s *ServiceManager) StartServers(context context.Context) error {
	for address, value := range s.GRPCServers {
		grpcServer := value.Server
		listener := value.Listener
		s.logger.Info().Msgf("Starting %s GRPC server", address)
		s.errGroup.Go(func() error {
			return grpcServer.Serve(listener)
		})

		httpServer := value.Gateway
		if httpServer.Server != nil {
			s.errGroup.Go(func() error {
				s.logger.Info().Msgf("Starting %s Gateway server", httpServer.Server.Addr)
				if httpServer.Certs == nil || httpServer.Certs.TLSCertPath == "" {
					err := httpServer.Server.ListenAndServe()
					if err != nil {
						return err
					}
				}
				if httpServer.Certs.TLSCertPath != "" {
					err := httpServer.Server.ListenAndServeTLS(httpServer.Certs.TLSCertPath, httpServer.Certs.TLSKeyPath)
					if err != nil {
						return err
					}
				}
				return nil
			})
		}
	}
	return nil
}

func (s *ServiceManager) StopServers(context context.Context) {
	for address, value := range s.GRPCServers {
		s.logger.Info().Msgf("Stopping %s GRPC server", address)
		value.Server.GracefulStop()
	}
}
