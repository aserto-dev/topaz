package resolvers

import (
	"context"

	runtime "github.com/aserto-dev/runtime"
)

type RuntimeResolver interface {
	RuntimeFromContext(ctx context.Context, policyName string) (*runtime.Runtime, error)
	GetRuntime(ctx context.Context, tenantID, policyName string) (*runtime.Runtime, error)
}
