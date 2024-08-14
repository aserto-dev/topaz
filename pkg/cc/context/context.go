package context

import (
	"context"

	"github.com/aserto-dev/topaz/pkg/cc/signals"

	"golang.org/x/sync/errgroup"
)

// ErrGroupAndContext wraps a context and an error group.
type ErrGroupAndContext struct {
	Ctx      context.Context
	ErrGroup *errgroup.Group
}

// NewContext creates a context that responds to user signals.
func NewContext() *ErrGroupAndContext {
	errGroup, ctx := errgroup.WithContext(signals.SetupSignalHandler())

	return &ErrGroupAndContext{
		Ctx:      ctx,
		ErrGroup: errGroup,
	}
}

// NewTestContext creates a context that can be used for testing.
func NewTestContext() *ErrGroupAndContext {
	errGroup, ctx := errgroup.WithContext(context.Background())

	return &ErrGroupAndContext{
		Ctx:      ctx,
		ErrGroup: errGroup,
	}
}
