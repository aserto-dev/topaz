package resolvers

import (
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	dsa1 "github.com/authzen/access.go/api/access/v1"
)

type DirectoryResolver interface {
	GetDS() dsr3.ReaderClient
	GetAuthZen() dsa1.AccessClient
}
