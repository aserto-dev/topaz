package builder

import (
	"context"
	"net"
	"reflect"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

type ServiceManager struct {
	Context  context.Context
	logger   *zerolog.Logger
	errGroup *errgroup.Group

	Servers       map[string]*Server
	DependencyMap map[string][]string
}

func NewServiceManager(logger *zerolog.Logger) *ServiceManager {

	serviceLogger := logger.With().Str("component", "service-manager").Logger()
	errGroup, ctx := errgroup.WithContext(context.Background())
	return &ServiceManager{
		Context:       ctx,
		logger:        &serviceLogger,
		Servers:       make(map[string]*Server),
		DependencyMap: make(map[string][]string),
		errGroup:      errGroup,
	}
}

func (s *ServiceManager) AddGRPCServer(server *Server) error {
	s.Servers[server.Config.GRPC.ListenAddress] = server
	return nil
}

func (s *ServiceManager) StartServers(ctx context.Context) error {
	for serverAddress, value := range s.Servers {
		address := serverAddress
		serverDetails := value

		// log all service details.
		s.logDetails(address, &serverDetails.Config.GRPC)
		s.logDetails(address, &serverDetails.Config.Gateway)
		s.logDetails(address, &serverDetails.Config.Health)

		s.errGroup.Go(func() error {
			if dependesOnArray, ok := s.DependencyMap[address]; ok {
				for _, dependesOn := range dependesOnArray {
					s.logger.Info().Msgf("%s waiting for %s", address, dependesOn)
					<-s.Servers[dependesOn].Started // wait for started from the dependenent service.
				}
			}
			grpcServer := serverDetails.Server
			listener := serverDetails.Listener
			s.logger.Info().Msgf("Starting %s GRPC server", address)
			s.errGroup.Go(func() error {
				return grpcServer.Serve(listener)
			})

			httpServer := serverDetails.Gateway
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
			if serverDetails.Health != nil {
				healthServer := serverDetails.Health
				healthListener, err := net.Listen("tcp", serverDetails.Config.Health.ListenAddress)
				s.logger.Info().Msgf("Starting %s Health server", serverDetails.Config.Health.ListenAddress)
				if err != nil {
					return err
				}
				s.errGroup.Go(func() error {
					return healthServer.GRPCServer.Serve(healthListener)
				})
			}

			serverDetails.Started <- true // send started information.
			return nil
		})
	}
	return nil
}

func (s *ServiceManager) logDetails(address string, element interface{}) {
	ref := reflect.ValueOf(element).Elem()
	typeOfT := ref.Type()

	for i := 0; i < ref.NumField(); i++ {
		f := ref.Field(i)
		s.logger.Debug().Str("address", address).Msgf("%s = %v\n", typeOfT.Field(i).Name, f.Interface())
	}
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
