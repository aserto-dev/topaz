package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/aserto-dev/certs"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

const (
	svcName = "authorizer"
)

type registeredServer struct {
	name        string
	start, stop func(ctx context.Context) error
}

type Server struct {
	ctx      context.Context
	cfg      *config.Common
	errGroup *errgroup.Group
	logger   *zerolog.Logger

	grpcServer        *grpc.Server
	grpcServerOptions []grpc.ServerOption
	GRPCRegistrations GRPCRegistrations
	gtwServer         *http.Server
	healthServer      *HealthServer

	gtwMux               *runtime.ServeMux
	HandlerRegistrations HandlerRegistrations

	registeredServers []registeredServer
}

func NewServer(
	ctx context.Context,
	logger *zerolog.Logger,
	cfg *config.Common,
	errGroup *errgroup.Group,
	grpcRegistrations GRPCRegistrations,
	handlerRegistrations HandlerRegistrations,
	gtwServer *http.Server,
	gtwMux *runtime.ServeMux,

) (*Server, func(), error) {

	newLogger := logger.With().Str("component", "api.edge-server").Logger()

	healthServer := newGRPCHealthServer()

	server := &Server{
		ctx:                  ctx,
		cfg:                  cfg,
		logger:               &newLogger,
		HandlerRegistrations: handlerRegistrations,
		GRPCRegistrations:    grpcRegistrations,
		gtwServer:            gtwServer,
		healthServer:         healthServer,
		gtwMux:               gtwMux,
		errGroup:             errGroup,
	}

	stopFunc := func() {
		err := server.Stop()
		if err != nil {
			logger.Error().Err(err).Msg("failed to stop server")
		}
	}

	return server, stopFunc, nil
}

// Adds Server Options to the GRPC server, for example GRPC Unary and Stream middlewares.
func (s *Server) AddGRPCServerOptions(grpcOptions ...grpc.ServerOption) {
	s.grpcServerOptions = append(s.grpcServerOptions, grpcOptions...)
}

// Registers additional servers to the app, for example, metrics server.
func (s *Server) RegisterServer(name string, start, stop func(ctx context.Context) error) {
	s.registeredServers = append(s.registeredServers, registeredServer{
		name:  name,
		start: start,
		stop:  stop,
	})
}

func (s *Server) Start(ctx context.Context) error {
	s.logger.Info().Msg("server::Start")

	grpc.EnableTracing = true

	// Start additional servers first.
	for _, regServer := range s.registeredServers {
		regSrv := regServer
		s.errGroup.Go(func() error {
			err := regSrv.start(ctx)
			if err != nil {
				return errors.Wrapf(err, "failed to start [%s]", regSrv.name)
			}
			return nil
		})
	}

	if err := s.startHealthServer(); err != nil {
		return err
	}

	if err := s.startGRPCServer(); err != nil {
		return err
	}

	if err := s.startGatewayServer(); err != nil {
		return err
	}

	s.healthServer.Server.SetServingStatus(fmt.Sprintf("grpc.health.v1.%s", svcName), healthpb.HealthCheckResponse_SERVING)

	return nil
}

// Stop stops the GRPC and HTTP servers, as well as their health servers.
func (s *Server) Stop() error {
	var result error

	s.logger.Info().Msg("Server stopping.")

	ctx, shutdownCancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer shutdownCancel()

	if s.gtwServer != nil {
		err := s.gtwServer.Shutdown(ctx)
		if err != nil {
			if err == context.Canceled {
				s.logger.Info().Msg("server context was canceled - shutting down")
			} else {
				result = multierror.Append(result, errors.Wrap(err, "failed to stop gateway server"))
			}
		}
	}

	// Stop additional servers.
	for _, registeredServer := range s.registeredServers {
		err := registeredServer.stop(ctx)
		if err != nil {
			return errors.Wrapf(err, "failed to stop server [%s]", registeredServer.name)
		}
	}

	if s.healthServer != nil {
		s.healthServer.Server.SetServingStatus(
			fmt.Sprintf("grpc.health.v1.%s", svcName),
			healthpb.HealthCheckResponse_NOT_SERVING,
		)
	}

	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}

	if s.healthServer.GRPCServer != nil {
		s.healthServer.GRPCServer.GracefulStop()
	}

	err := s.errGroup.Wait()
	if err != nil {
		s.logger.Info().Err(err).Msg("shutdown complete")
	}

	return result
}

// registerGateway registers the gateway server with a _running_ gRPC server.
func (s *Server) registerGateway() error {
	_, port, err := net.SplitHostPort(s.cfg.API.GRPC.ListenAddress)
	if err != nil {
		return errors.Wrap(err, "failed to determine port from configured GRPC listen address")
	}

	dialAddr := fmt.Sprintf("dns:///127.0.0.1:%s", port)

	tlsCreds, err := certs.GatewayAsClientTLSCreds(s.cfg.API.GRPC.Certs)
	if err != nil {
		return errors.Wrap(err, "failed to calculate tls config for gateway service")
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(tlsCreds),
		grpc.WithBlock(),
	}

	err = s.HandlerRegistrations(s.ctx, s.gtwMux, dialAddr, opts)
	if err != nil {
		return errors.Wrap(err, "failed to register handlers with the gateway")
	}

	return nil
}

func (s *Server) startHealthServer() error {
	healthListener, err := net.Listen("tcp", s.cfg.API.Health.ListenAddress)
	if err != nil {
		s.logger.Error().Err(err).Str("address", s.cfg.API.Health.ListenAddress).Msg("grpc health socket failed to listen")
		return errors.Wrap(err, "grpc health socket failed to listen")
	}

	s.logger.Info().Str("address", s.cfg.API.Health.ListenAddress).Msg("GRPC Health Server starting")
	s.errGroup.Go(func() error {
		return s.healthServer.GRPCServer.Serve(healthListener)
	})

	return nil
}

func (s *Server) startGRPCServer() error {
	s.logger.Info().Str("address", s.cfg.API.GRPC.ListenAddress).Msg("GRPC Server starting")
	grpcListener, err := net.Listen("tcp", s.cfg.API.GRPC.ListenAddress)
	if err != nil {
		return errors.Wrap(err, "grpc socket failed to listen")
	}

	grpcServer, err := newGRPCServer(s.cfg, s.logger, s.GRPCRegistrations, s.grpcServerOptions...)
	if err != nil {
		return err
	}
	s.grpcServer = grpcServer

	s.errGroup.Go(func() error {
		err := s.grpcServer.Serve(grpcListener)
		if err != nil {
			s.logger.Error().Err(err).Str("address", s.cfg.API.GRPC.ListenAddress).Msg("GRPC Server failed to listen")
		}
		return errors.Wrap(err, "grpc server failed to listen")
	})

	return nil
}

func (s *Server) startGatewayServer() error {
	s.logger.Info().Msg("Registering gRPC-Gateway handlers")
	if err := s.registerGateway(); err != nil {
		return errors.Wrap(err, "failed to register grpc gateway handlers")
	}

	s.errGroup.Go(func() error {
		if s.cfg.API.Gateway.HTTP {
			s.logger.Info().Str("address", "http://"+s.cfg.API.Gateway.ListenAddress).Msg("gRPC-Gateway endpoint starting")
			return s.gtwServer.ListenAndServe()
		}
		s.logger.Info().Str("address", "https://"+s.cfg.API.Gateway.ListenAddress).Msg("gRPC-Gateway endpoint starting")
		return s.gtwServer.ListenAndServeTLS("", "")
	})

	return nil
}
