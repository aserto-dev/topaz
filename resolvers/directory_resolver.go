package resolvers

import (
	"context"

	dsr2 "github.com/aserto-dev/go-directory/aserto/directory/reader/v2"
)

type DirectoryResolver interface {
	GetDS(ctx context.Context) (dsr2.ReaderClient, error)
}
