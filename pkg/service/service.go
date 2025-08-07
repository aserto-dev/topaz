package service

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type Service interface {
	RegisterGRPC(server *grpc.Server)
	RegisterGateway(ctx context.Context, mux *runtime.ServeMux, addr string, opts []grpc.DialOption) error
}

type (
	GRPCRegistrar    func(server *grpc.Server)
	GatewayRegistrar func(ctx context.Context, mux *runtime.ServeMux, addr string, opts []grpc.DialOption) error
)

func NoGateway(_ context.Context, _ *runtime.ServeMux, _ string, _ []grpc.DialOption) error {
	return nil
}

type Impl struct {
	regGRPC    GRPCRegistrar
	regGateway GatewayRegistrar
}

var Noop = NewImpl(func(_ *grpc.Server) {}, NoGateway)

func NewImpl(g GRPCRegistrar, gw GatewayRegistrar) *Impl {
	return &Impl{g, gw}
}

func (s *Impl) RegisterGRPC(server *grpc.Server) {
	s.regGRPC(server)
}

func (s *Impl) RegisterGateway(ctx context.Context, mux *runtime.ServeMux, addr string, opts []grpc.DialOption) error {
	return s.regGateway(ctx, mux, addr, opts)
}
