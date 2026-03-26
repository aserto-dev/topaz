package grpcutil

import (
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

type Middleware interface {
	Unary() grpc.UnaryServerInterceptor
	Stream() grpc.StreamServerInterceptor
}

type Middlewares []Middleware

func (m Middlewares) AsGRPCOptions() (grpc.ServerOption, grpc.ServerOption) {
	unaryInterceptors := []grpc.UnaryServerInterceptor{}
	streamInterceptors := []grpc.StreamServerInterceptor{}

	for _, middleware := range m {
		if middleware == nil {
			continue
		}

		unaryInterceptor := middleware.Unary()
		streamInterceptor := middleware.Stream()

		if unaryInterceptor != nil {
			unaryInterceptors = append(unaryInterceptors, unaryInterceptor)
		}

		if streamInterceptor != nil {
			streamInterceptors = append(streamInterceptors, streamInterceptor)
		}
	}

	return grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(unaryInterceptors...)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(streamInterceptors...))
}
