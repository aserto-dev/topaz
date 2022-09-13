package server

import (
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/aserto-dev/aserto-grpc/grpcutil"
	"github.com/aserto-dev/go-utils/certs"
	"github.com/aserto-dev/topaz/pkg/cc/config"
)

func newGRPCServer(cfg *config.Common, logger *zerolog.Logger, registrations GRPCRegistrations, middlewares grpcutil.Middlewares) (*grpc.Server, error) {
	grpc.EnableTracing = true

	if err := view.Register(ocgrpc.DefaultServerViews...); err != nil {
		logger.Error().Err(err).Msg("failed to register ocgrpc server views")
	}

	connectionTimeout := time.Duration(cfg.API.GRPC.ConnectionTimeoutSeconds) * time.Second
	tlsCreds, err := certs.GRPCServerTLSCreds(cfg.API.GRPC.Certs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to calculate tls config")
	}

	tlsAuth := grpc.Creds(tlsCreds)

	unaryMiddlewares, streamMiddlewares := middlewares.AsGRPCOptions()

	server := grpc.NewServer(
		tlsAuth,
		grpc.ConnectionTimeout(connectionTimeout),
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
		unaryMiddlewares,
		streamMiddlewares,
	)

	reflection.Register(server)

	registrations(server)

	return server, nil
}
