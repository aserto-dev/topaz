package bdb

import (
	"context"
	"encoding/json"

	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

type Message[T any] interface {
	proto.Message
	UnmarshalVT(b []byte) error
	*T
}

var (
	marshalOpts = proto.MarshalOptions{
		AllowPartial:  false,
		Deterministic: false,
		UseCachedSize: false,
	}
	unmarshalOpts = proto.UnmarshalOptions{
		Merge:          false,
		AllowPartial:   false,
		DiscardUnknown: true,
	}
)

func marshal[T any, M Message[T]](t M) ([]byte, error) {
	return marshalOpts.Marshal(t)
}

func unmarshal[T any, M Message[T]](b []byte) (M, error) {
	var t T

	msg := M(&t)
	if err := unmarshalOpts.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	return &t, nil
}

func unmarshalTo[T any, M Message[T]](b []byte, dst M) error {
	return dst.UnmarshalVT(b)
}

func Get[T any, M Message[T]](ctx context.Context, tx *bolt.Tx, path Path, key []byte) (M, error) {
	buf, err := GetKey(tx, path, key)
	if err != nil {
		return nil, err
	}

	return unmarshal[T, M](buf)
}

func List[T any, M Message[T]](ctx context.Context, tx *bolt.Tx, path Path) ([]M, error) {
	result := []M{}

	b, err := SetBucket(tx, path)
	if err != nil {
		return result, err
	}

	c := b.Cursor()

	for key, value := c.First(); key != nil; key, value = c.Next() {
		i, err := unmarshal[T, M](value)
		if err != nil {
			return []M{}, err
		}

		result = append(result, i)
	}

	return result, nil
}

func Set[T any, M Message[T]](ctx context.Context, tx *bolt.Tx, path Path, key []byte, t M) (M, error) {
	buf, err := marshal(t)
	if err != nil {
		return nil, err
	}

	if err := SetKey(tx, path, key, buf); err != nil {
		return nil, err
	}

	return t, nil
}

func Delete(ctx context.Context, tx *bolt.Tx, path Path, key []byte) error {
	return DeleteKey(tx, path, key)
}

func marshalAny[T any](v T) ([]byte, error) {
	return json.Marshal(&v)
}

func unmarshalAny[T any](buf []byte) (*T, error) {
	var t T
	if err := json.Unmarshal(buf, &t); err != nil {
		return nil, err
	}

	return &t, nil
}

func GetAny[T any](ctx context.Context, tx *bolt.Tx, path Path, key []byte) (*T, error) {
	buf, err := GetKey(tx, path, key)
	if err != nil {
		return nil, err
	}

	return unmarshalAny[T](buf)
}

func SetAny[T any](ctx context.Context, tx *bolt.Tx, path Path, key []byte, t *T) (*T, error) {
	buf, err := marshalAny(t)
	if err != nil {
		return nil, err
	}

	if err := SetKey(tx, path, key, buf); err != nil {
		return nil, err
	}

	return t, nil
}
