package resolvers

import (
	"context"

	runtime "github.com/aserto-dev/runtime"
)

type RuntimeResolver interface {
	RuntimeFromContext(ctx context.Context, policyName, instanceLabel string) (*runtime.Runtime, error)
	GetRuntime(ctx context.Context, tenantID, policyName, instanceLabel string) (*runtime.Runtime, error)
	PeekRuntime(ctx context.Context, tenantID, policyName, instanceLabel string) (*runtime.Runtime, error)
	ReloadRuntime(ctx context.Context, tenantID, policyName, instanceLabel string) error
	ListRuntimes(ctx context.Context) (map[string]*runtime.Runtime, error)
	UnloadRuntime(ctx context.Context, tenantID, policyName, instanceLabel string)
}
