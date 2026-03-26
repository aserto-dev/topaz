package tests_test

import (
	"context"

	"google.golang.org/grpc"
)

var (
	UnaryInfo = &grpc.UnaryServerInfo{
		FullMethod: "TestService.UnaryMethod",
	}

	StreamInfo = &grpc.StreamServerInfo{
		FullMethod:     "TestService.StreamMethod",
		IsClientStream: false,
		IsServerStream: true,
	}
)

type Handler struct {
	output any
	err    error
}

func NewHandler(output any, err error) *Handler {
	return &Handler{output, err}
}

func (h *Handler) Unary(ctx context.Context, req any) (any, error) {
	return h.output, h.err
}

func (h *Handler) Stream(srv any, stream grpc.ServerStream) error {
	return h.err
}
