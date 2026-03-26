package request

import (
	"context"
	"strings"

	grpcutil "github.com/aserto-dev/topaz/internal/grpc"
	"github.com/aserto-dev/topaz/internal/header"
	"github.com/google/uuid"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type RequestIDMiddleware struct{}

func NewRequestIDMiddleware() *RequestIDMiddleware {
	return &RequestIDMiddleware{}
}

var _ grpcutil.Middleware = &RequestIDMiddleware{}

// Unary returns a new unary server interceptor that creates a request ID
// and sets it on the context.
// If the request already contains a proper request ID, it will be persisted, and a new
// request ID will be appended to it, separated by a dot '.'.
// In order to chain request IDs (when calling other services), the caller can extract the ID
// from the context, or use metadata.AppendToOutgoingContext with the context available in
// GRPC handlers.
func (m *RequestIDMiddleware) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		id, err := m.requestID(ctx)
		if err != nil {
			return nil, err
		}

		newCtx := header.ContextWithRequestID(ctx, id)

		if id != "" {
			err = grpc.SetHeader(newCtx, metadata.Pairs(string(header.HeaderAsertoRequestID), id))
			if err != nil {
				return nil, err
			}
		}

		result, err := handler(newCtx, req)

		return result, err
	}
}

// Stream returns a new stream server interceptor that creates a request ID
// and sets it on the context.
// If the request already contains a proper request ID, it will be persisted, and a new
// request ID will be appended to it, separated by a dot '.'.
// In order to chain request IDs (when calling other services), the caller can extract the ID
// from the context, or use metadata.AppendToOutgoingContext with the context available in
// GRPC handlers.
func (m *RequestIDMiddleware) Stream() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()

		id, err := m.requestID(ctx)
		if err != nil {
			return err
		}

		newCtx := header.ContextWithRequestID(ctx, id)

		if id != "" {
			err = grpc.SetHeader(newCtx, metadata.Pairs(string(header.HeaderAsertoRequestID), id))
			if err != nil {
				return err
			}
		}

		wrapped := grpcmiddleware.WrapServerStream(stream)
		wrapped.WrappedContext = newCtx

		return handler(srv, wrapped)
	}
}

// UnaryClient returns a new unary client interceptor that forwards request IDs to the outgoing context.
// If the context doesn't contain a request ID, no ID is added to the outgoing context.
func (m *RequestIDMiddleware) UnaryClient() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		id := header.ExtractRequestID(ctx)
		if id != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, string(header.HeaderAsertoRequestID), id)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StreamClient returns a new stream client interceptor that forwards request IDs to the outgoing context.
// If the context doesn't contain a request ID, no ID is added to the outgoing context.
func (m *RequestIDMiddleware) StreamClient() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		id := header.ExtractRequestID(ctx)
		if id != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, string(header.HeaderAsertoRequestID), id)
		}

		return streamer(ctx, desc, cc, method, opts...)
	}
}

func (m *RequestIDMiddleware) requestID(ctx context.Context) (string, error) {
	reqid, err := uuid.NewUUID()
	if err != nil {
		return "", errors.Wrap(err, "failed to generate new request id")
	}

	incomingID := IncomingRequestID(ctx)
	if incomingID != "" {
		incomingID = strings.Split(incomingID, ".")[0]
		if header.IsValidUUID(incomingID) {
			return incomingID + "." + reqid.String(), nil
		}

		log.Debug().Err(err).Msg("invalid request id")
	}

	return reqid.String(), nil
}

func IncomingRequestID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	return requestIDFromMetadata(md)
}

func OutgoingRequestID(ctx context.Context) string {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return ""
	}

	return requestIDFromMetadata(md)
}

func requestIDFromMetadata(md metadata.MD) string {
	headerMetaData, ok := md[string(header.HeaderAsertoRequestID)]
	if !ok || len(headerMetaData) == 0 {
		headerMetaData, ok = md[strings.ToLower(string(header.HeaderAsertoRequestID))]

		if !ok || len(headerMetaData) == 0 {
			return ""
		}
	}

	return headerMetaData[0]
}
