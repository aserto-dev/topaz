package resolvers

import (
	"context"

	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
)

type DirectoryResolver interface {
	GetDS(ctx context.Context) (dsr3.ReaderClient, error)
}
