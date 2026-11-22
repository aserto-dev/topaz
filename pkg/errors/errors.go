package errors

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	MessageKey = "msg"
	colon      = ": "
)

var (
	ErrUnknown = NewAsertoError("E00000", codes.Internal, http.StatusInternalServerError, "an unknown error has occurred")

	asertoErrors = make(map[string]*AsertoError) //nolint:gochecknoglobals
)

// AsertoError represents a well known error
// coming from an Aserto service.
type AsertoError struct {
	Code     string
	GRPCCode codes.Code
	HTTPCode int
	Message  string
	data     map[string]string
	wrapped  error
}

func NewAsertoError(code string, grpcCode codes.Code, httpCode int, msg string) *AsertoError {
	asertoError := &AsertoError{
		Code:     code,
		GRPCCode: grpcCode,
		HTTPCode: httpCode,
		Message:  msg,
		data:     map[string]string{},
		wrapped:  nil,
	}
	asertoErrors[code] = asertoError

	return asertoError
}

func (e *AsertoError) Data() map[string]string {
	return e.Copy().data
}

func (e *AsertoError) Copy() *AsertoError {
	dataCopy := make(map[string]string, len(e.data))

	maps.Copy(dataCopy, e.data)

	return &AsertoError{
		Code:     e.Code,
		GRPCCode: e.GRPCCode,
		HTTPCode: e.HTTPCode,
		Message:  e.Message,
		data:     dataCopy,
		wrapped:  e.wrapped,
	}
}

func (e *AsertoError) Error() string {
	innerMessage := ""

	if e.wrapped != nil {
		innerMessage = e.wrapped.Error()
	}

	if msg, ok := e.data[MessageKey]; ok {
		if innerMessage != "" {
			innerMessage = colon + innerMessage
		}

		innerMessage = msg + innerMessage
	}

	if innerMessage == "" {
		return fmt.Sprintf("%s %s", e.Code, e.Message)
	}

	return fmt.Sprintf("%s %s: %s", e.Code, e.Message, innerMessage)
}

func (e *AsertoError) Fields() map[string]any {
	result := make(map[string]any, len(e.data))

	var aerr *AsertoError
	if errors.As(e.wrapped, &aerr) {
		maps.Copy(result, aerr.Fields())
	}

	for k, v := range e.data {
		result[k] = v
	}

	return result
}

// Err associates err with the AsertoError, replacing any error previously
// associated via Err.
func (e *AsertoError) Err(err error) *AsertoError {
	if err == nil {
		return e
	}

	c := e.Copy()
	c.wrapped = err

	return c
}

func (e *AsertoError) Msg(message string) *AsertoError {
	c := e.Copy()

	if message != "" {
		if existingMsg, ok := c.data[MessageKey]; ok {
			c.data[MessageKey] = existingMsg + colon + message
		} else {
			c.data[MessageKey] = message
		}
	}

	return c
}

func (e *AsertoError) Msgf(message string, args ...any) *AsertoError {
	c := e.Copy()

	message = fmt.Sprintf(message, args...)

	if existingMsg, ok := c.data[MessageKey]; ok {
		c.data[MessageKey] = existingMsg + colon + message
	} else {
		c.data[MessageKey] = message
	}

	return c
}

func (e *AsertoError) Str(key, value string) *AsertoError {
	c := e.Copy()
	c.data[key] = value

	return c
}

func (e *AsertoError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.wrapped
}

func (e *AsertoError) Cause() error {
	return e.wrapped
}

func (e *AsertoError) MarshalZerologObject(event *zerolog.Event) {
	event.Str("error", e.Error())
	event.Fields(e.Fields())
}

// ErrorInfo returns the errdetails.ErrorInfo for this error, with HTTPCode
// folded into Metadata under HTTPStatusErrorMetadata so it survives the
// grpc-gateway boundary (see CustomErrorHandler).
func (e *AsertoError) ErrorInfo(reason string) *errdetails.ErrorInfo {
	data := e.Data()
	data[HTTPStatusErrorMetadata] = strconv.Itoa(e.HTTPCode)

	return &errdetails.ErrorInfo{
		Reason:   reason,
		Metadata: data,
		Domain:   e.Code,
	}
}

func (e *AsertoError) GRPCStatus() *status.Status {
	errResult := status.New(e.GRPCCode, e.Message)

	errResult, err := errResult.WithDetails(e.ErrorInfo(""))
	if err != nil {
		return status.New(codes.Internal, "internal failure setting up error details, please contact the administrator")
	}

	return errResult
}

func (e *AsertoError) Ctx(ctx context.Context) error {
	return WithContext(e, ctx)
}

// FromGRPCStatus returns an Aserto error based on a given grpcStatus. The details that are not of type errdetails.ErrorInfo are dropped.
// and if there are details from multiple errors, the aserto error will be constructed based on the first one.
func FromGRPCStatus(grpcStatus status.Status) *AsertoError {
	if len(grpcStatus.Details()) == 0 {
		return ErrUnknown.Msg(grpcStatus.Message())
	}

	for _, detail := range grpcStatus.Details() {
		t, ok := detail.(*errdetails.ErrorInfo)
		if !ok {
			continue
		}

		registered := asertoErrors[t.GetDomain()]
		if registered == nil {
			return nil
		}

		result := registered.Copy()
		result.data = t.GetMetadata()

		return result
	}

	return nil
}

// Logger retrieves the most inner logger associated with an error.
func Logger(err error) *zerolog.Logger {
	var (
		logger *zerolog.Logger
		ce     *ContextError
	)

	if err == nil {
		return logger
	}

	for {
		if errors.As(err, &ce) {
			if ctxLogger := extractLogger(ce.Ctx); ctxLogger != nil {
				logger = ctxLogger
			}
		}

		err = errors.Unwrap(err)
		if err == nil {
			break
		}
	}

	return logger
}

func UnwrapAsertoError(err error) *AsertoError {
	if err == nil {
		return nil
	}

	initialError := errors.Cause(err)
	if initialError == nil {
		initialError = err
	}

	// try to process Aserto error.
	for {
		var aErr *AsertoError
		if ok := errors.As(err, &aErr); ok {
			return aErr
		}

		err = errors.Unwrap(err)
		if err == nil {
			break
		}
	}

	// If it's not an Aserto error, try to construct one from grpc status.
	grpcStatus, ok := status.FromError(initialError)
	if ok {
		aErr := FromGRPCStatus(*grpcStatus)
		if aErr != nil {
			return aErr
		}
	}

	return nil
}

func CodeToAsertoError(code string) *AsertoError {
	return asertoErrors[code]
}

/**
 * Retrieve the logger associated with the context using zerolog.Ctx(ctx).
 * If the retrieved logger is either the default context logger or has a disabled level, it returns nil.
 */
func extractLogger(ctx context.Context) *zerolog.Logger {
	if ctx == nil {
		return nil
	}

	logger := zerolog.Ctx(ctx)
	if logger == zerolog.DefaultContextLogger || logger.GetLevel() == zerolog.Disabled {
		logger = nil
	}

	return logger
}
