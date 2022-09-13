package resolvers

import (
	"context"

	ds2 "github.com/aserto-dev/go-directory/aserto/directory/v2"
	"github.com/aserto-dev/topaz/directory"
)

type DirectoryResolver interface {
	GetDS(ctx context.Context) (ds2.DirectoryClient, error)
	DirectoryFromContext(ctx context.Context) (directory.Directory, error)
	GetDirectory(ctx context.Context, instanceID string) (directory.Directory, error)
	ReloadDirectory(ctx context.Context, instanceID string) error
	ListDirectories(ctx context.Context) ([]string, error)
	RemoveDirectory(ctx context.Context, instanceID string) error
}
