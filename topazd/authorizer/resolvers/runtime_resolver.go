package resolvers

import (
	"context"

	runtime "github.com/aserto-dev/runtime"
)

type RuntimeResolver interface {
	GetRuntime(ctx context.Context) (*runtime.Runtime, error)
}
