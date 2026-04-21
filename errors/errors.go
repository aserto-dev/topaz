package errors

import (
	"context"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	Code       string
	StatusCode codes.Code
	Message    string
	HTTPCode   int
	data       map[string]string
	errs       []error
}

func NewAsertoError(code string, statusCode codes.Code, httpCode int, msg string) *AsertoError {
	asertoError := &AsertoError{code, statusCode, msg, httpCode, map[string]string{}, nil}
	asertoErrors[code] = asertoError

	return asertoError
}

func (e *AsertoError) Data() map[string]string {
	return e.Copy().data
}

// SameAs returns true if the provided error is an AsertoError
// and has the same error code.
func (e *AsertoError) SameAs(err error) bool {
	var aErr *AsertoError
	if ok := errors.As(err, &aErr); err == nil || !ok {
		return false
	}

	return aErr.Code == e.Code
}

func (e *AsertoError) Copy() *AsertoError {
	dataCopy := make(map[string]string, len(e.data))

	maps.Copy(dataCopy, e.data)

	return &AsertoError{
		Code:       e.Code,
		StatusCode: e.StatusCode,
		Message:    e.Message,
		data:       dataCopy,
		errs:       e.errs,
		HTTPCode:   e.HTTPCode,
	}
}

func (e *AsertoError) Error() string {
	errsMessage := ""

	if len(e.errs) > 0 {
		errsMessage = e.errs[0].Error()

		for _, err := range e.errs[1:] {
			errsMessage = errsMessage + colon + err.Error()
		}
	}

	innerMessage := errsMessage

	if len(e.data) > 0 {
		for k, v := range e.data {
			if k == "msg" {
				if innerMessage != "" {
					innerMessage = colon + innerMessage
				}

				innerMessage = v + innerMessage
			}
		}
	}

	if innerMessage == "" {
		return fmt.Sprintf("%s %s", e.Code, e.Message)
	}

	return fmt.Sprintf("%s %s: %s", e.Code, e.Message, innerMessage)
}

func (e *AsertoError) Fields() map[string]any {
	result := make(map[string]any, len(e.data))

	for _, err := range e.errs {
		var aerr *AsertoError
		if ok := errors.As(err, &aerr); ok {
			maps.Copy(result, aerr.Fields())
		}
	}

	for k, v := range e.data {
		result[k] = v
	}

	return result
}

// Err associates err with the AsertoError.
func (e *AsertoError) Err(err error) *AsertoError {
	if err == nil {
		return e
	}

	c := e.Copy()
	c.errs = append(c.errs, err)

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

func (e *AsertoError) Int(key string, value int) *AsertoError {
	c := e.Copy()
	c.data[key] = strconv.Itoa(value)

	return c
}

func (e *AsertoError) Int32(key string, value int32) *AsertoError {
	c := e.Copy()
	c.data[key] = strconv.FormatInt(int64(value), 10)

	return c
}

func (e *AsertoError) Int64(key string, value int64) *AsertoError {
	c := e.Copy()
	c.data[key] = strconv.FormatInt(value, 10)

	return c
}

func (e *AsertoError) Bool(key string, value bool) *AsertoError {
	c := e.Copy()
	c.data[key] = strconv.FormatBool(value)

	return c
}

func (e *AsertoError) Duration(key string, value time.Duration) *AsertoError {
	c := e.Copy()
	c.data[key] = value.String()

	return c
}

func (e *AsertoError) Time(key string, value time.Time) *AsertoError {
	c := e.Copy()
	c.data[key] = value.UTC().Format(time.RFC3339)

	return c
}

func (e *AsertoError) FromReader(key string, value io.Reader) *AsertoError {
	buf := &strings.Builder{}

	if _, err := io.Copy(buf, value); err != nil {
		return e.Err(err)
	}

	c := e.Copy()
	c.data[key] = buf.String()

	return c
}

func (e *AsertoError) Interface(key string, value any) *AsertoError {
	c := e.Copy()
	c.data[key] = fmt.Sprintf("%+v", value)

	return c
}

func (e *AsertoError) Unwrap() error {
	if e == nil {
		return nil
	}

	if len(e.errs) > 0 {
		return e.errs[len(e.errs)-1]
	}

	return nil
}

func (e *AsertoError) Cause() error {
	if len(e.errs) > 0 {
		return e.errs[len(e.errs)-1]
	}

	return nil
}

func (e *AsertoError) MarshalZerologObject(event *zerolog.Event) {
	event.Str("error", e.Error())
	event.Fields(e.Fields())
}

func (e *AsertoError) GRPCStatus() *status.Status {
	errResult := status.New(e.StatusCode, e.Message)

	errResult, err := errResult.WithDetails(&errdetails.ErrorInfo{
		Metadata: e.Data(),
		Domain:   e.Code,
	})
	if err != nil {
		return status.New(codes.Internal, "internal failure setting up error details, please contact the administrator")
	}

	return errResult
}

func (e *AsertoError) WithGRPCStatus(grpcCode codes.Code) *AsertoError {
	c := e.Copy()
	c.StatusCode = grpcCode

	return c
}

func (e *AsertoError) WithHTTPStatus(httpStatus int) *AsertoError {
	c := e.Copy()
	c.HTTPCode = httpStatus

	return c
}

func (e *AsertoError) Ctx(ctx context.Context) error {
	return WithContext(e, ctx)
}

// FromGRPCStatus returns an Aserto error based on a given grpcStatus. The details that are not of type errdetails.ErrorInfo are dropped.
// and if there are details from multiple errors, the aserto error will be constructed based on the first one.
func FromGRPCStatus(grpcStatus status.Status) *AsertoError {
	var result *AsertoError

	if len(grpcStatus.Details()) == 0 {
		return ErrUnknown.Msg(grpcStatus.Message())
	}

	for _, detail := range grpcStatus.Details() {
		if t, ok := detail.(*errdetails.ErrorInfo); ok {
			result = asertoErrors[t.GetDomain()]
			if result == nil {
				return nil
			}

			result.data = t.GetMetadata()
		}

		if result != nil {
			break
		}
	}

	return result
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

// Equals returns true if the given errors are Aserto errors with the same code or both of them are nil.
func Equals(err1, err2 error) bool {
	asertoErr1 := UnwrapAsertoError(err1)
	asertoErr2 := UnwrapAsertoError(err2)

	if err1 == nil && err2 == nil {
		return true
	}

	if asertoErr1 == nil || asertoErr2 == nil {
		return false
	}

	return asertoErr1.Code == asertoErr2.Code
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
