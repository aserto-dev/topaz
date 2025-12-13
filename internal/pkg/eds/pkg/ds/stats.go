package ds

import (
	"context"
	"sync/atomic"

	"github.com/aserto-dev/azm/model"
	"github.com/aserto-dev/azm/stats"
	dsc3 "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	"github.com/aserto-dev/topaz/internal/pkg/eds/pkg/bdb"

	bolt "go.etcd.io/bbolt"
)

// CalculateStats returns a Stats object with the counts of all objects and relations.
func CalculateStats(ctx context.Context, tx *bolt.Tx) (*stats.Stats, error) {
	s := NewStats()

	if err := s.CountObjects(ctx, tx); err != nil {
		return nil, err
	}

	if err := s.CountRelations(ctx, tx); err != nil {
		return nil, err
	}

	return s.Stats, nil
}

type Stats struct {
	*stats.Stats
}

func NewStats() *Stats {
	return &Stats{stats.NewStats()}
}

func (s *Stats) CountObjects(ctx context.Context, tx *bolt.Tx) error {
	iter, err := bdb.NewScanIterator[dsc3.Object](ctx, tx, bdb.ObjectsPath)
	if err != nil {
		return err
	}

	for iter.Next() {
		obj := iter.Value()
		s.incObject(obj)
	}

	return nil
}

func (s *Stats) CountRelations(ctx context.Context, tx *bolt.Tx) error {
	iter, err := bdb.NewScanIterator[dsc3.Relation](ctx, tx, bdb.RelationsObjPath)
	if err != nil {
		return err
	}

	for iter.Next() {
		rel := iter.Value()
		s.incRelation(rel)
	}

	return nil
}

func (s *Stats) incObject(obj *dsc3.Object) {
	ot, ok := s.ObjectTypes[model.ObjectName(obj.GetType())]
	if !ok {
		ot = &stats.ObjectType{
			Relations: stats.Relations{},
		}
		s.ObjectTypes[model.ObjectName(obj.GetType())] = ot
	}

	atomic.AddInt32(&ot.ObjCount, 1)
}

func (s *Stats) incRelation(rel *dsc3.Relation) {
	objType := model.ObjectName(rel.GetObjectType())
	relation := model.RelationName(rel.GetRelation())
	subType := model.ObjectName(rel.GetSubjectType())
	subRel := model.RelationName(rel.GetSubjectRelation())

	if rel.GetSubjectId() == "*" {
		subType += ":*"
	}

	// object_types
	ot := s.ObjectTypes[objType]
	if ot == nil {
		ot = &stats.ObjectType{
			Relations: stats.Relations{},
		}
		s.ObjectTypes[objType] = ot
	}

	atomic.AddInt32(&ot.Count, 1)

	// relations
	re := ot.Relations[relation]
	if re == nil {
		re = &stats.Relation{
			SubjectTypes: stats.SubjectTypes{},
		}
		ot.Relations[relation] = re
	}

	atomic.AddInt32(&re.Count, 1)

	// subject_types
	st := re.SubjectTypes[subType]
	if st == nil {
		st = &stats.SubjectType{
			SubjectRelations: stats.SubjectRelations{},
		}
		re.SubjectTypes[subType] = st
	}

	atomic.AddInt32(&st.Count, 1)

	// subject_relations
	if subRel != "" {
		sr := st.SubjectRelations[subRel]
		if sr == nil {
			sr = &stats.SubjectRelation{}
			st.SubjectRelations[subRel] = sr
		}

		atomic.AddInt32(&sr.Count, 1)
	}
}
