package context

import (
	"context"

	"github.com/aserto-dev/topaz/topazd/signals"

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
