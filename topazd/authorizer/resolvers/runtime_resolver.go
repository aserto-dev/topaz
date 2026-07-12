package resolvers

import (
	"context"

	runtime "github.com/aserto-dev/topaz/internal/runtime"
)

type RuntimeResolver interface {
	GetRuntime(ctx context.Context) (*runtime.Runtime, error)
}
