package datasync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/bdb"

	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type watermark struct {
	LastUpdated   string                 `json:"last_updated"`
	Timestamp     *timestamppb.Timestamp `json:"ts"`
	TotalCount    uint                   `json:"count,omitempty"`
	ObjectCount   uint                   `json:"obj_count,omitempty"`
	RelationCount uint                   `json:"rel_count,omitempty"`
}

func newWatermark() *watermark {
	return &watermark{
		LastUpdated:   "",
		Timestamp:     &timestamppb.Timestamp{Seconds: 0, Nanos: 0},
		TotalCount:    0,
		ObjectCount:   0,
		RelationCount: 0,
	}
}

// getFilterSize determine the number of entries for the cuckoo filter.
// a filter size of 1 mln entries results in a ~= 2Mib memory allocation
// the default and minimum filterSize configure is 100K entries.
func (wm *watermark) getFilterSize() uint {
	const (
		initFilterSize uint = 100000
		growthFactor   uint = 2
		rounding       uint = 1000
	)

	// when TotalCount is zero, implying uninitialized or unknown we use default filter size (default 100K).
	if wm.TotalCount == 0 {
		return initFilterSize
	}

	// determine filter minimum capacity, allowing for growths, based on growthFactor constant (default 2x).
	growthSize := wm.TotalCount * growthFactor

	// if growth size is smaller than initial filter size, use the initial size.
	if growthSize < initFilterSize {
		return initFilterSize
	}

	// round up to the nearest upper boundary (default 1000).
	remainder := growthSize % rounding

	filterSize := growthSize + (rounding - remainder)

	return filterSize
}

func maxTS(lhs, rhs *timestamppb.Timestamp) *timestamppb.Timestamp {
	if lhs.GetSeconds() > rhs.GetSeconds() {
		return lhs
	} else if lhs.GetSeconds() == rhs.GetSeconds() && lhs.GetNanos() > rhs.GetNanos() {
		return lhs
	}

	return rhs
}

func (s *Sync) getWatermark() *watermark {
	r, err := os.Open(s.syncFilename())
	if err != nil {
		return newWatermark()
	}
	defer r.Close()

	var wm watermark

	dec := json.NewDecoder(r)
	if err := dec.Decode(&wm); err != nil {
		return newWatermark()
	}

	return &wm
}

func (s *Sync) setWatermark(ts *timestamppb.Timestamp) error {
	if ts == nil {
		panic("ts is nil")
	}

	newTS := maxTS(s.getWatermark().Timestamp, ts)

	wm := newWatermark()
	wm.Timestamp = newTS
	wm.LastUpdated = newTS.AsTime().Format(time.RFC3339Nano)

	objStats, err := s.dbStats(bdb.ObjectsPath)
	if err != nil {
		return err
	}

	relStats, err := s.dbStats(bdb.RelationsObjPath)
	if err != nil {
		return err
	}

	wm.ObjectCount = uint(objStats.KeyN)   //nolint:gosec // G115: integer overflow conversion int -> uint
	wm.RelationCount = uint(relStats.KeyN) //nolint:gosec // G115: integer overflow conversion int -> uint

	wm.TotalCount = wm.ObjectCount + wm.RelationCount

	w, err := os.Create(s.syncFilename())
	if err != nil {
		return err
	}
	defer w.Close()

	if err := json.NewEncoder(w).Encode(wm); err != nil {
		return err
	}

	_ = w.Sync() // flush sync watermark.

	return nil
}

func (s *Sync) syncFilename() string {
	dir, file := filepath.Split(s.Client.store.DB().Path())
	return filepath.Join(dir, fmt.Sprintf("%s.%s", file, "sync"))
}

func (s *Sync) dbStats(path bdb.Path) (bolt.BucketStats, error) {
	var stats bolt.BucketStats

	err := s.store.DB().View(func(tx *bolt.Tx) error {
		stats = bucketStats(tx, path)
		return nil
	})

	return stats, err
}

func bucketStats(tx *bolt.Tx, path bdb.Path) bolt.BucketStats {
	b, err := bdb.SetBucket(tx, path)
	if err != nil {
		return bolt.BucketStats{}
	}

	return b.Stats()
}
