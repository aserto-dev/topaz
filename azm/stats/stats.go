package stats

import (
	"github.com/aserto-dev/topaz/azm/model"
)

type Stats struct {
	ObjectTypes ObjectTypes `json:"object_types,omitempty"`
}

func NewStats() *Stats {
	return &Stats{
		ObjectTypes: make(ObjectTypes),
	}
}

type ObjectTypes map[model.ObjectName]*ObjectType

type ObjectType struct {
	ObjCount  int32     `json:"_obj_count,omitempty"`
	Count     int32     `json:"_count,omitempty"`
	Relations Relations `json:"relations,omitempty"`
}

type Relations map[model.RelationName]*Relation

type Relation struct {
	Count        int32        `json:"_count,omitempty"`
	SubjectTypes SubjectTypes `json:"subject_types,omitempty"`
}

type SubjectTypes map[model.ObjectName]*SubjectType

type SubjectType struct {
	Count            int32            `json:"_count,omitempty"`
	SubjectRelations SubjectRelations `json:"subject_relations,omitempty"`
}

type SubjectRelations map[model.RelationName]*SubjectRelation

type SubjectRelation struct {
	Count int32 `json:"_count,omitempty"`
}

func (s *Stats) ObjectRefCount(on model.ObjectName) int32 {
	if ot, ok := s.ObjectTypes[on]; ok {
		return ot.Count + ot.ObjCount
	}

	return 0
}

func (s *Stats) RelationRefCount(on model.ObjectName, rn model.RelationName) int32 {
	if ot, ok := s.ObjectTypes[on]; ok {
		if rt, ok := ot.Relations[rn]; ok {
			return rt.Count
		}
	}

	return 0
}

func (s *Stats) RelationSubjectCount(on model.ObjectName, rn model.RelationName, sn model.ObjectName, sr model.RelationName) int32 {
	ot, ok := s.ObjectTypes[on]
	if !ok {
		return 0
	}

	rt, ok := ot.Relations[rn]
	if !ok {
		return 0
	}

	st, ok := rt.SubjectTypes[sn]
	if !ok {
		return 0
	}

	switch {
	case sr != "":
		if srt, ok := st.SubjectRelations[sr]; ok {
			return srt.Count
		}
	default:
		subjCount := int32(0)
		for _, subj := range st.SubjectRelations {
			subjCount += subj.Count
		}

		return st.Count - subjCount
	}

	return 0
}
