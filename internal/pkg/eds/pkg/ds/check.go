package ds

import (
	"context"

	"github.com/aserto-dev/azm/cache"
	"github.com/aserto-dev/azm/graph"
	"github.com/aserto-dev/azm/safe"
	dsc3 "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/go-directory/pkg/derr"
	"github.com/aserto-dev/go-directory/pkg/prop"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/bdb"

	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/types/known/structpb"
)

type check struct {
	*safe.SafeCheck
}

func Check(i *dsr3.CheckRequest) *check {
	return &check{safe.Check(i)}
}

func (i *check) Exec(ctx context.Context, tx *bolt.Tx, mc *cache.Cache) (*dsr3.CheckResponse, error) {
	if err := i.RelationIdentifiersExist(ctx, tx); err != nil {
		return &dsr3.CheckResponse{
			Check:   false,
			Context: SetContextWithReason(err),
		}, err
	}

	return mc.Check(i.CheckRequest, getRelations(ctx, tx))
}

func getRelations(ctx context.Context, tx *bolt.Tx) graph.RelationReader {
	return func(r *dsc3.RelationIdentifier, pool graph.RelationPool, out *[]*dsc3.RelationIdentifier) error {
		keyFilter := RelationIdentifierBuffer()
		defer ReturnRelationIdentifierBuffer(keyFilter)

		path, valueFilter := RelationIdentifier(r).Filter(keyFilter)

		return bdb.ScanWithFilter(ctx, tx, path, keyFilter.Bytes(), valueFilter, pool, out)
	}
}

func (i *check) RelationIdentifiersExist(ctx context.Context, tx *bolt.Tx) error {
	if !i.relationIdentifierExist(
		ctx, tx, bdb.RelationsSubPath,
		ObjectIdentifier(&dsc3.ObjectIdentifier{ObjectType: i.SubjectType, ObjectId: i.SubjectId}).Key(),
	) {
		return derr.ErrObjectNotFound.Msgf("subject %s:%s", i.SubjectType, i.SubjectId)
	}

	if !i.relationIdentifierExist(
		ctx, tx, bdb.RelationsObjPath,
		ObjectIdentifier(&dsc3.ObjectIdentifier{ObjectType: i.ObjectType, ObjectId: i.ObjectId}).Key(),
	) {
		return derr.ErrObjectNotFound.Msgf("object %s:%s", i.ObjectType, i.ObjectId)
	}

	return nil
}

func (i *check) relationIdentifierExist(ctx context.Context, tx *bolt.Tx, path bdb.Path, keyFilter []byte) bool {
	exists, err := bdb.KeyPrefixExists[dsc3.Relation](ctx, tx, path, keyFilter)
	if err != nil {
		return false
	}

	return exists
}

func SetContextWithReason(err error) *structpb.Struct {
	return &structpb.Struct{
		Fields: map[string]*structpb.Value{
			prop.Reason: structpb.NewStringValue(err.Error()),
		},
	}
}
