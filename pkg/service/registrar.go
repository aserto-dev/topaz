package service

import (
	"context"

	gorilla "github.com/gorilla/mux"
	"google.golang.org/grpc"

	"github.com/aserto-dev/topaz/pkg/servers"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type Registrar interface {
	RegisterGRPC(server *grpc.Server)
	RegisterGateway(ctx context.Context, mux *runtime.ServeMux, addr string, opts []grpc.DialOption) error
	RegisterHTTP(ctx context.Context, cfg *servers.HTTPServer, router *gorilla.Router) error
}

type (
	GRPCRegistrar    func(server *grpc.Server)
	GatewayRegistrar func(ctx context.Context, mux *runtime.ServeMux, addr string, opts []grpc.DialOption) error
	HTTPRegistrar    func(ctx context.Context, cfg *servers.HTTPServer, router *gorilla.Router) error
)

func NoGateway(_ context.Context, _ *runtime.ServeMux, _ string, _ []grpc.DialOption) error {
	return nil
}

func NoHTTP(_ context.Context, _ *servers.HTTPServer, _ *gorilla.Router) error { return nil }

type Impl struct {
	regGRPC    GRPCRegistrar
	regGateway GatewayRegistrar
	regHTTP    HTTPRegistrar
}

var Noop = NewImpl(func(_ *grpc.Server) {}, NoGateway, NoHTTP)

func NewImpl(g GRPCRegistrar, gw GatewayRegistrar, h HTTPRegistrar) *Impl {
	return &Impl{g, gw, h}
}

func (s *Impl) RegisterGRPC(server *grpc.Server) {
	s.regGRPC(server)
}

func (s *Impl) RegisterGateway(ctx context.Context, mux *runtime.ServeMux, addr string, opts []grpc.DialOption) error {
	return s.regGateway(ctx, mux, addr, opts)
}

func (s *Impl) RegisterHTTP(ctx context.Context, cfg *servers.HTTPServer, router *gorilla.Router) error {
	return s.regHTTP(ctx, cfg, router)
}
