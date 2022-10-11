package resolvers

import (
	"context"

	ds2 "github.com/aserto-dev/go-directory/aserto/directory/v2"
)

type DirectoryResolver interface {
	GetDS(ctx context.Context) (ds2.DirectoryClient, error)
}
