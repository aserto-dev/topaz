package x

import (
	"context"
	"slices"

	"github.com/hashicorp/go-multierror"
)

type closeFunc func(context.Context) error

type Closer []closeFunc

func (c Closer) Close(ctx context.Context) error {
	var errs error

	for _, close := range slices.Backward(c) {
		if err := close(ctx); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs
}

func CloserFunc(f func()) closeFunc {
	return func(_ context.Context) error {
		f()
		return nil
	}
}

func CloserErr(f func() error) closeFunc {
	return func(_ context.Context) error {
		return f()
	}
}
