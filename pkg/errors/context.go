package errors

import (
	"context"

	"github.com/pkg/errors"
)

// ContextError represents a standard error
// that can also encapsulate a context.
type ContextError struct {
	Err error
	Ctx context.Context //nolint:containedctx
}

func WithContext(err error, ctx context.Context) *ContextError {
	return &ContextError{
		Err: err,
		Ctx: ctx,
	}
}

func WrapContext(err error, ctx context.Context, message string) *ContextError {
	return WithContext(errors.Wrap(err, message), ctx)
}

func WrapfContext(err error, ctx context.Context, format string, args ...any) *ContextError {
	return WithContext(errors.Wrapf(err, format, args...), ctx)
}

func (ce *ContextError) Error() string {
	return ce.Err.Error()
}

func (ce *ContextError) Cause() error {
	return errors.Cause(ce.Unwrap())
}

func (ce *ContextError) Unwrap() error {
	return ce.Err
}
