package ds

import (
	"context"
	"time"

	dsc3 "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/bdb"

	bolt "go.etcd.io/bbolt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func UpdateMetadataObject(ctx context.Context, tx *bolt.Tx, path []string, keyFilter []byte, msg *dsc3.Object) (*dsc3.Object, error) {
	// get timestamp once for transaction.
	ts := timestamppb.New(time.Now().UTC())

	// get current instance.
	cur, err := bdb.Get[dsc3.Object](ctx, tx, path, keyFilter)

	switch {
	case status.Code(err) == codes.NotFound:
		// new instance, set created_at timestamp.
		msg.CreatedAt = ts
		// if new instance set Etag to empty string.
		msg.Etag = ""

	case err != nil:
		return nil, err
	default:
		// existing instance, propagate created_at timestamp.
		msg.CreatedAt = cur.GetCreatedAt()
	}

	// always set updated_at timestamp.
	msg.UpdatedAt = ts

	if cur.GetEtag() != "" {
		msg.Etag = cur.GetEtag()
	}

	return msg, nil
}

func UpdateMetadataRelation(ctx context.Context, tx *bolt.Tx, path []string, key []byte, msg *dsc3.Relation) (*dsc3.Relation, error) {
	// get timestamp once for transaction.
	ts := timestamppb.New(time.Now().UTC())

	// get current instance.
	cur, err := bdb.Get[dsc3.Relation](ctx, tx, path, key)

	switch {
	case status.Code(err) == codes.NotFound:
		// new instance, set created_at timestamp.
		msg.CreatedAt = ts
		// if new instance set Etag to empty string.
		msg.Etag = ""

	case err != nil:
		return nil, err
	default:
		// existing instance, propagate created_at timestamp.
		msg.CreatedAt = cur.GetCreatedAt()
	}

	// always set updated_at timestamp.
	msg.UpdatedAt = ts

	if cur.GetEtag() != "" {
		msg.Etag = cur.GetEtag()
	}

	return msg, nil
}
