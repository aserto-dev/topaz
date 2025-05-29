package middleware

import (
	"context"
	"time"

	"github.com/google/uuid"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/rs/zerolog"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	aerr "github.com/aserto-dev/errors"
)

type Logging struct {
	logger *zerolog.Logger
}

func NewLogging(logger *zerolog.Logger) *Logging {
	return &Logging{logger}
}

func (m *Logging) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		method, _ := grpc.Method(ctx)
		logger := m.logger.With().Str("method", method).Interface("request", req).Logger()
		ctx = logger.WithContext(ctx)

		logger.Trace().Msg("grpc call start")

		start := time.Now()
		result, err := handler(ctx, req)

		logger.Trace().Dur("duration-ms", time.Since(start)).Msg("grpc call end")

		return result, handleError(&logger, err)
	}
}

func (m *Logging) Stream() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()
		method, _ := grpc.Method(ctx)
		logger := m.logger.With().Str("method", method).Logger()
		ctx = logger.WithContext(ctx)

		logger.Trace().Msg("grpc stream call")

		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = ctx

		return handleError(&logger, handler(srv, wrapped))
	}
}

func handleError(logger *zerolog.Logger, rpcErr error) error {
	if rpcErr == nil {
		return nil
	}

	if errorLogger := aerr.Logger(rpcErr); errorLogger != nil {
		logger = errorLogger
	}

	errID, _ := uuid.NewUUID() // NewUUID never returns an error.

	asertoErr := aerr.UnwrapAsertoError(rpcErr)
	if asertoErr == nil {
		asertoErr = aerr.ErrUnknown
	}

	asertoErr = asertoErr.Int(aerr.HTTPStatusErrorMetadata, asertoErr.HTTPCode)

	logger.Warn().Stack().Err(rpcErr).
		Str("error-id", errID.String()).
		Str("error-code", asertoErr.Code).
		Int("status-code", int(asertoErr.StatusCode)).
		Fields(asertoErr.Fields()).
		Msg(asertoErr.Message)

	rpcStatus, err := status.New(asertoErr.StatusCode, asertoErr.Error()).
		WithDetails(&errdetails.ErrorInfo{
			Reason:   errID.String(),
			Metadata: asertoErr.Data(),
			Domain:   asertoErr.Code,
		})
	if err != nil {
		logger.Error().Err(rpcErr).Err(err).Msg("failed to create grpc status for error")
		return status.New(codes.Internal, "internal failure setting up error details, please contact the administrator").Err()
	}

	return rpcStatus.Err()
}
