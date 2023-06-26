package builder

import (
	"context"
	"net"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

type ServiceManager struct {
	Context  context.Context
	logger   *zerolog.Logger
	errGroup *errgroup.Group

	Servers map[string]*Server
}

func NewServiceManager(logger *zerolog.Logger) *ServiceManager {
	errGroup, ctx := errgroup.WithContext(context.Background())
	return &ServiceManager{
		Context:  ctx,
		logger:   logger,
		Servers:  make(map[string]*Server),
		errGroup: errGroup,
	}
}

func (s *ServiceManager) AddGRPCServer(server *Server) error {
	s.Servers[server.Config.GRPC.ListenAddress] = server
	return nil
}

func (s *ServiceManager) StartServers(ctx context.Context) error {
	for address, value := range s.Servers {
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
		if value.Health != nil {
			healthServer := value.Health
			healthListener, err := net.Listen("tcp", value.Config.Health.ListenAddress)
			s.logger.Info().Msgf("Starting %s Health server", value.Config.Health.ListenAddress)
			if err != nil {
				return err
			}
			s.errGroup.Go(func() error {
				return healthServer.GRPCServer.Serve(healthListener)
			})
		}
	}
	return nil
}

func (s *ServiceManager) StopServers(ctx context.Context) {
	for address, value := range s.Servers {
		s.logger.Info().Msgf("Stopping %s GRPC server", address)
		value.Server.GracefulStop()
		if value.Gateway.Server != nil {
			err := value.Gateway.Server.Shutdown(ctx)
			if err != nil {
				s.logger.Err(err).Msgf("failed to shutdown gateway for %s", address)
			}
		}
		if value.Health != nil && value.Health.GRPCServer != nil {
			value.Health.GRPCServer.GracefulStop()
		}
	}
}
