package gerr

import (
	"context"

	aerr "github.com/aserto-dev/topaz/errors"
	grpcutil "github.com/aserto-dev/topaz/internal/grpc"
	"github.com/google/uuid"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorMiddleware struct{}

func NewErrorMiddleware() *ErrorMiddleware {
	return &ErrorMiddleware{}
}

var _ grpcutil.Middleware = &ErrorMiddleware{}

func (m *ErrorMiddleware) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		result, handlerErr := handler(ctx, req)
		return result, m.handleError(ctx, handlerErr)
	}
}

func (m *ErrorMiddleware) Stream() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()

		wrapped := grpcmiddleware.WrapServerStream(stream)
		wrapped.WrappedContext = ctx

		handlerErr := handler(srv, wrapped)

		return m.handleError(ctx, handlerErr)
	}
}

func (m *ErrorMiddleware) handleError(ctx context.Context, handlerErr error) error {
	if handlerErr == nil {
		return nil
	}

	log := zerolog.Ctx(ctx)
	if errorLogger := aerr.Logger(handlerErr); errorLogger != nil {
		log = errorLogger
	}

	if log == nil {
		zlog.Error().Msgf("ERROR - ZEROLOG LOGGER MISSING FROM CONTEXT: %v\n", handlerErr)
		return status.New(codes.Internal, "internal logging error, please contact the administrator").Err()
	}

	errID, err := uuid.NewUUID()
	if err != nil {
		log.Error().Err(handlerErr).Err(err).Msg("failed to create error id")
		return status.New(codes.Internal, "internal failure to generate an error id, please contact the administrator").Err()
	}

	asertoErr := aerr.UnwrapAsertoError(handlerErr)
	if asertoErr == nil {
		asertoErr = aerr.ErrUnknown
	}

	asertoErr = asertoErr.Int(aerr.HTTPStatusErrorMetadata, asertoErr.HTTPCode)

	log.Warn().Stack().Err(handlerErr).
		Ctx(ctx).
		Str("error-id", errID.String()).
		Str("error-code", asertoErr.Code).
		Int("status-code", int(asertoErr.StatusCode)).
		Fields(asertoErr.Fields()).
		Msg(asertoErr.Message)

	errResult := status.New(asertoErr.StatusCode, asertoErr.Error())

	errResult, err = errResult.WithDetails(&errdetails.ErrorInfo{
		Reason:   errID.String(),
		Metadata: asertoErr.Data(),
		Domain:   asertoErr.Code,
	})
	if err != nil {
		log.Error().Err(handlerErr).Err(err).Msg("failed to setup error result")
		return status.New(codes.Internal, "internal failure setting up error details, please contact the administrator").Err()
	}

	return errResult.Err()
}
