package mempool

import (
	dsc "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RelationsPool = CollectionPool[*dsc.RelationIdentifier]

func NewRelationsPool() *RelationsPool {
	return NewCollectionPool[*dsc.RelationIdentifier](NewRelationAllocator())
}

type RelationAllocator struct {
	tsPool *Pool[*timestamppb.Timestamp]
}

func NewRelationAllocator() *RelationAllocator {
	return &RelationAllocator{
		tsPool: NewPool[*timestamppb.Timestamp](
			func() *timestamppb.Timestamp {
				return new(timestamppb.Timestamp)
			}),
	}
}

func (ra *RelationAllocator) New() *dsc.RelationIdentifier {
	rel := dsc.RelationIdentifierFromVTPool()
	return rel
}

func (ra *RelationAllocator) Reset(rel *dsc.RelationIdentifier) {
	rel.ReturnToVTPool()
}
