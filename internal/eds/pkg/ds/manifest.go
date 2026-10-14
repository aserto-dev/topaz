package ds

import (
	"hash/fnv"
	"strconv"

	dsc "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
)

type manifest struct {
	*dsc.Manifest
}

func Manifest(man *dsc.Manifest) *manifest {
	return &manifest{man}
}

// Get, hydrates the manifest from the _manifest bucket
// _metadata/{name}/{version}/metadata
// _metadata/{name}/{version}/body.
// func (m *manifest) Get(ctx context.Context, tx *bolt.Tx) (*manifest, error) {
// 	return Manifest(&dsc.Manifest{}), nil
// }

// GetModel, hydrates the model cache from the _manifest
// _metadata/{name}/{version}/model.
// func (m *manifest) GetModel(ctx context.Context, tx *bolt.Tx) (*model.Model, error) {
// 	if ok, _ := bdb.BucketExists(tx, bdb.ManifestPath); !ok {
// 		return nil, bdb.ErrPathNotFound
// 	}

// 	mod, err := bdb.GetAny[model.Model](ctx, tx, bdb.ManifestPath, bdb.ModelKey)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return mod, nil
// }

// Set, persists the manifest body in the _manifest bucket
// _metadata/{name}/{version}/metadata
// _metadata/{name}/{version}/body.
// func (m *manifest) Set(ctx context.Context, tx *bolt.Tx, buf *bytes.Buffer) error {
// if _, err := bdb.CreateBucket(tx, bdb.ManifestPathV3); err != nil {
// 	return err
// }

// if _, err := bdb.Set(ctx, tx, bdb.ManifestPathV3, bdb.ManifestKeyV3, m.Metadata); err != nil {
// 	return err
// }

// m.Body = &dsm.Body{Data: buf.Bytes()}
// if _, err := bdb.Set(ctx, tx, bdb.ManifestPathV3, bdb.ManifestKeyV3, m.Body); err != nil {
// 	return err
// }

// 	return nil
// }

// SetModel, persists the model cache in the _manifest bucket
// _metadata/{name}/{version}/model.
// func (m *manifest) SetModel(ctx context.Context, tx *bolt.Tx, mod *model.Model) error {
// if mod.Metadata == nil {
// 	mod.Metadata = &model.Metadata{}
// }

// mod.Metadata.ETag = m.Metadata.GetEtag()
// mod.Metadata.UpdatedAt = m.Metadata.GetUpdatedAt().AsTime()

// if _, err := bdb.SetAny(ctx, tx, bdb.ManifestPath, bdb.ModelKey, mod); err != nil {
// 	return err
// }

// 	return nil
// }

// Delete
//
// !!! NOTE: delete manifest is a destructive operation !!!
//
// sets the manifest to an empty manifest,
// updates the model accordingly,
// deletes and recreates the objects and relations buckets.
// func (m *manifest) Delete(ctx context.Context, tx *bolt.Tx) error {
// if err := bdb.DeleteBucket(tx, bdb.ManifestPath); err != nil {
// 	return err
// }

// if _, err := bdb.CreateBucket(tx, bdb.ManifestPath); err != nil {
// 	return err
// }

// if err := bdb.DeleteBucket(tx, bdb.ObjectsPath); err != nil {
// 	return err
// }

// if _, err := bdb.CreateBucket(tx, bdb.ObjectsPath); err != nil {
// 	return err
// }

// if err := bdb.DeleteBucket(tx, bdb.RelationsObjPath); err != nil {
// 	return err
// }

// if _, err := bdb.CreateBucket(tx, bdb.RelationsObjPath); err != nil {
// 	return err
// }

// if err := bdb.DeleteBucket(tx, bdb.RelationsSubPath); err != nil {
// 	return err
// }

// if _, err := bdb.CreateBucket(tx, bdb.RelationsSubPath); err != nil {
// 	return err
// }

// 	return nil
// }

func (m *manifest) Hash() string {
	h := fnv.New64a()

	h.Reset()

	if _, err := h.Write(m.GetContent()); err != nil {
		return DefaultHash
	}

	return strconv.FormatUint(h.Sum64(), 10)
}
