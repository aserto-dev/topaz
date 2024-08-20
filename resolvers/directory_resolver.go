package resolvers

import (
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
)

type DirectoryResolver interface {
	GetDS() dsr3.ReaderClient
}
