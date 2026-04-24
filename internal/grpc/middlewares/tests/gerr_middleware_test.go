package tests_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/aserto-dev/logger"
	aerr "github.com/aserto-dev/topaz/errors"
	"github.com/aserto-dev/topaz/internal/grpc/middlewares/gerr"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errTest = aerr.NewAsertoError("X00000", codes.Internal, http.StatusInternalServerError, "a test error")

func TestUnaryServerWithWrappedError(t *testing.T) {
	assert := require.New(t)
	handler := NewHandler("output", errors.Wrap(errTest, "unimportant error"))

	ctx := grpc.NewContextWithServerTransportStream(
		RequestIDContext(t),
		ServerTransportStream(""),
	)
	_, err := gerr.NewErrorMiddleware().Unary()(ctx, "xyz", UnaryInfo, handler.Unary)
	assert.Error(err)
	assert.Contains(err.Error(), "a test error")
}

func TestUnaryServerWithFields(t *testing.T) {
	assert := require.New(t)
	handler := NewHandler(
		"output",
		errors.Wrap(errTest.Str("my-field", "deadbeef"), "another error"),
	)

	buf := bytes.NewBufferString("")
	testLogger := logger.TestLogger(buf)

	ctx := grpc.NewContextWithServerTransportStream(
		RequestIDContext(t),
		ServerTransportStream(""),
	)

	_, err := gerr.NewErrorMiddleware().Unary()(testLogger.WithContext(ctx), "xyz", UnaryInfo, handler.Unary)
	assert.Error(err)

	logOutput := buf.String()

	assert.Contains(logOutput, "deadbeef")
}

func TestUnaryServerWithDoubleCerr(t *testing.T) {
	assert := require.New(t)
	handler := NewHandler(
		"output",
		aerr.ErrUnknown.Err(errTest.Str("my-field", "deadbeef").Msg("old message")).Msg("new message"),
	)

	buf := bytes.NewBufferString("")
	testLogger := logger.TestLogger(buf)

	ctx := grpc.NewContextWithServerTransportStream(
		RequestIDContext(t),
		ServerTransportStream(""),
	)

	_, err := gerr.NewErrorMiddleware().Unary()(testLogger.WithContext(ctx), "xyz", UnaryInfo, handler.Unary)
	assert.Error(err)

	logOutput := buf.String()

	assert.Contains(logOutput, "new message")
	assert.Contains(logOutput, "deadbeef")
}

func TestSimpleInnerError(t *testing.T) {
	assert := require.New(t)
	handler := NewHandler("output", aerr.ErrUnknown.Err(errors.New("deadbeef")).Msg("failed to setup initial tag"))

	buf := bytes.NewBufferString("")
	testLogger := logger.TestLogger(buf)

	ctx := grpc.NewContextWithServerTransportStream(
		RequestIDContext(t),
		ServerTransportStream(""),
	)

	_, err := gerr.NewErrorMiddleware().Unary()(testLogger.WithContext(ctx), "xyz", UnaryInfo, handler.Unary)
	assert.Error(err)

	logOutput := buf.String()

	assert.Contains(logOutput, "deadbeef")
}

func TestDirectResult(t *testing.T) {
	assert := require.New(t)
	handler := NewHandler(
		"output",
		aerr.ErrUnknown.Err(errTest).Msg("failed to setup initial tag"),
	)

	buf := bytes.NewBufferString("")
	testLogger := logger.TestLogger(buf)

	ctx := grpc.NewContextWithServerTransportStream(
		RequestIDContext(t),
		ServerTransportStream(""),
	)

	_, err := gerr.NewErrorMiddleware().Unary()(testLogger.WithContext(ctx), "xyz", UnaryInfo, handler.Unary)
	assert.Error(err)

	s := status.Convert(err)
	errDetailsFound := false

	for _, detail := range s.Details() {
		switch t := detail.(type) {
		case *errdetails.ErrorInfo:
			errDetailsFound = true

			assert.Contains(t.GetMetadata(), "msg")
			assert.Contains(t.GetMetadata()["msg"], "failed to setup")
		}
	}

	assert.True(errDetailsFound)
	assert.Contains(s.Message(), "an unknown error has occurred")
	assert.Contains(s.Message(), "failed to setup initial tag")
	assert.Contains(err.Error(), "failed to setup initial tag")
}
