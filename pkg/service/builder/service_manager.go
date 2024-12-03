package builder

import (
	"context"
	"net"
	"net/http"
	"reflect"
	"time"

	"github.com/aserto-dev/go-aserto"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	go_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type ServiceManager struct {
	Context  context.Context
	logger   *zerolog.Logger
	errGroup *errgroup.Group

	Servers         map[string]*Service
	DependencyMap   map[string][]string
	HealthServer    *Health
	MetricServer    *http.Server
	shutdownTimeout int // timeout to force stop services in seconds
}

var reg *prometheus.Registry

func NewServiceManager(logger *zerolog.Logger) *ServiceManager {
	if reg == nil {
		reg = prometheus.NewRegistry()
	}
	serviceLogger := logger.With().Str("component", "service-manager").Logger()
	errGroup, ctx := errgroup.WithContext(context.Background())
	return &ServiceManager{
		Context:         ctx,
		logger:          &serviceLogger,
		Servers:         make(map[string]*Service),
		DependencyMap:   make(map[string][]string),
		errGroup:        errGroup,
		shutdownTimeout: 30,
	}
}

func (s *ServiceManager) WithShutdownTimeout(seconds int) *ServiceManager {
	s.shutdownTimeout = seconds
	return s
}

func (s *ServiceManager) AddGRPCServer(server *Service) error {
	s.Servers[server.Config.GRPC.ListenAddress] = server
	return nil
}

func (s *ServiceManager) SetupHealthServer(address string, certCfg *aserto.TLSConfig) error {
	healthServer := newGRPCHealthServer(certCfg)
	healthServer.Address = address

	s.HealthServer = healthServer
	healthListener, err := net.Listen("tcp", address)
	s.logger.Info().Msgf("Starting %s health server", address)
	if err != nil {
		return err
	}
	s.errGroup.Go(func() error {
		return healthServer.GRPCServer.Serve(healthListener)
	})
	return nil
}

func (s *ServiceManager) SetupMetricsServer(address string, certCfg *aserto.TLSConfig, enableZpages bool) ([]grpc.ServerOption,
	error,
) {
	metric := http.Server{
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	s.MetricServer = &metric
	mux := http.NewServeMux()

	grpcm := grpc_prometheus.NewServerMetrics(
		grpc_prometheus.WithServerCounterOptions(),
		grpc_prometheus.WithServerHandlingTimeHistogram(),
	)

	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewBuildInfoCollector())
	reg.MustRegister(grpcm)

	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		Registry: reg,
	}))

	metric.Handler = mux
	metric.Addr = address

	s.logger.Info().Msgf("Starting %s metric server", address)

	if !certCfg.HasCert() {
		s.errGroup.Go(metric.ListenAndServe)
	} else {
		s.errGroup.Go(func() error {
			return metric.ListenAndServeTLS(certCfg.Cert, certCfg.Key)
		})
	}

	exemplarFromContext := func(ctx context.Context) prometheus.Labels {
		method, ok := grpc.Method(ctx)

		if ok {
			return prometheus.Labels{"method": method}
		}
		return nil
	}

	var opts []grpc.ServerOption

	unary := grpc.ChainUnaryInterceptor(grpcm.UnaryServerInterceptor(grpc_prometheus.WithExemplarFromContext(exemplarFromContext)))
	stream := grpc.ChainStreamInterceptor(grpcm.StreamServerInterceptor(grpc_prometheus.WithExemplarFromContext(exemplarFromContext)))
	opts = append(opts,
		unary,
		stream,
		grpc.ChainUnaryInterceptor(go_prometheus.UnaryServerInterceptor),
		grpc.ChainStreamInterceptor(go_prometheus.StreamServerInterceptor),
	)

	return opts, nil
}

func (s *ServiceManager) StartServers(ctx context.Context) error {
	for serverAddress, value := range s.Servers {
		address := serverAddress
		serverDetails := value

		// log all service details.
		s.logDetails(address, &serverDetails.Config.GRPC)
		s.logDetails(address, &serverDetails.Config.Gateway)

		s.errGroup.Go(func() error {
			if dependsOnArray, ok := s.DependencyMap[address]; ok {
				for _, dependsOn := range dependsOnArray {
					s.logger.Info().Msgf("%s waiting for %s", address, dependsOn)
					<-s.Servers[dependsOn].Started // wait for started from the dependent service.
				}
			}
			grpcServer := serverDetails.Server
			listener := serverDetails.Listener
			s.logger.Info().Msgf("Starting %s gRPC server", address)
			s.errGroup.Go(func() error {
				return grpcServer.Serve(listener)
			})

			if serverDetails.Gateway != nil {
				httpServer := serverDetails.Gateway
				if httpServer.Server != nil {
					s.errGroup.Go(func() error {
						s.logger.Info().Msgf("Starting %s gateway server", httpServer.Server.Addr)
						if httpServer.Certs.HasCert() {
							err := httpServer.Server.ListenAndServeTLS(httpServer.Certs.Cert, httpServer.Certs.Key)
							if err != nil {
								return err
							}
						} else {
							err := httpServer.Server.ListenAndServe()
							if err != nil {
								return err
							}
						}
						return nil
					})
				}
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

	for i := range ref.NumField() {
		f := ref.Field(i)
		s.logger.Debug().Str("address", address).Msgf("%s = %v\n", typeOfT.Field(i).Name, f.Interface())
	}
}

func (s *ServiceManager) StopServers(ctx context.Context) {
	timeout := time.Duration(s.shutdownTimeout) * time.Second
	timeoutContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if s.HealthServer != nil {
		s.logger.Info().Msgf("Stopping %s health server", s.HealthServer.Address)
		if !shutDown(s.HealthServer.GRPCServer, timeout) {
			s.logger.Warn().Msgf("Stopped %s health server forcefully", s.HealthServer.Address)
		}
	}

	if s.MetricServer != nil {
		s.logger.Info().Msgf("Stopping %s metric server", s.MetricServer.Addr)
		err := s.MetricServer.Shutdown(timeoutContext)
		if err != nil {
			s.logger.Err(err).Msgf("failed to shutdown metric server %s", s.MetricServer.Addr)
			s.logger.Debug().Msgf("forcefully closing metric server %s", s.MetricServer.Addr)
			if err := s.MetricServer.Close(); err != nil {
				s.logger.Err(err).Msgf("failed to close the metric server %s", s.MetricServer.Addr)
			}
		}
	}

	for address, value := range s.Servers {
		// shutdown gateway service first, as it is layered on-top of the gRPC service.
		if value.Gateway != nil && value.Gateway.Server != nil {
			s.logger.Info().Msgf("Stopping %s gateway server", value.Gateway.Server.Addr)
			err := value.Gateway.Server.Shutdown(timeoutContext)
			if err != nil {
				s.logger.Err(err).Msgf("failed to shutdown gateway for %s", address)
				s.logger.Debug().Msgf("forcefully closing gateway %s", address)
				if err := value.Gateway.Server.Close(); err != nil {
					s.logger.Err(err).Msgf("failed to close gateway server %s", address)
				}
			}
		}

		// shutdown gRPC service.
		s.logger.Info().Msgf("Stopping %s gRPC server", address)
		if !shutDown(value.Server, timeout) {
			s.logger.Warn().Msgf("Stopped %s gRPC forcefully", address)
		}

		// run cleanup routines.
		for _, cleanup := range value.Cleanup {
			s.logger.Info().Msgf("Running cleanups for %s", address)
			cleanup()
		}
	}
}

func shutDown(server *grpc.Server, timeout time.Duration) bool {
	result := make(chan bool, 1)
	go func() {
		server.GracefulStop()
		result <- true
	}()
	select {
	case <-time.After(timeout):
		server.Stop()
		return false
	case response := <-result:
		return response
	}
}
