package builder

import (
	"github.com/aserto-dev/topaz/topazd/middleware"
	"google.golang.org/grpc"
)

type middlewares struct {
	auth    middleware.Server
	logging *middleware.Logging
}

func (m *middlewares) unary() grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(m.logging.Unary(), m.auth.Unary())
}

func (m *middlewares) stream() grpc.ServerOption {
	return grpc.ChainStreamInterceptor(m.logging.Stream(), m.auth.Stream())
}
