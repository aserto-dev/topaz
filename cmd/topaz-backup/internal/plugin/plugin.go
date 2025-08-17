package plugin

import (
	"context"
	"os"
)

const ReadOnly os.FileMode = 0o400

type StorePlugin interface {
	Run(ctx context.Context) error
}
