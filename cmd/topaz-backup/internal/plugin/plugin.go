package plugin

import (
	"context"
)

type StorePlugin interface {
	Run(ctx context.Context) error
}
