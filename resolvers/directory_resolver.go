package resolvers

import (
	"google.golang.org/grpc"
)

type DirectoryResolver interface {
	GetConn() *grpc.ClientConn
}
