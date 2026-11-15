package resolvers

import (
	"context"

	"github.com/aserto-dev/topaz/internal/runtime"
)

type RuntimeResolver interface {
	GetRuntime(ctx context.Context) (*runtime.Runtime, error)
}
