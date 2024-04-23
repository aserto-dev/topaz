package resolvers

import (
	"context"

	"github.com/aserto-dev/go-directory/pkg/datasync"
)

type EdgeSyncResolver interface {
	GetDataSync(ctx context.Context) (datasync.Client, error)
}
