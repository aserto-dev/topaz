package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/aserto-dev/topaz/api/directory/pkg/derr"

	set "github.com/deckarep/golang-set/v2"
	"github.com/samber/lo"
)

const (
	ModelVersion int = 5

	ArrowSymbol    = "->"
	WildcardSymbol = "*"
)

type Model struct {
	Version  int                    `json:"version"`
	Objects  map[ObjectName]*Object `json:"types"`
	Metadata *Metadata              `json:"metadata"`
	inverted *Model                 `json:"-"`
}

type Metadata struct {
	UpdatedAt time.Time `json:"updated_at"`
	ETag      string    `json:"etag"`
}

func New(r io.Reader) (*Model, error) {
	m := Model{}
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&m); err != nil {
		return nil, err
	}

	m.inverted = newInverter(&m).invert()

	return &m, nil
}

type ObjectID string

func (id ObjectID) String() string {
	return string(id)
}

func (id ObjectID) IsWildcard() bool {
	return id == WildcardSymbol
}

type relation struct {
	on  ObjectName
	oid ObjectID
	rn  RelationName
	sn  ObjectName
	sid ObjectID
	srn RelationName
}

func (r *relation) String() string {
	srn := ""
	if r.srn != "" {
		srn = "#" + r.srn.String()
	}

	return fmt.Sprintf("%s:%s#%s@%s:%s%s", r.on, r.oid, r.rn, r.sn, r.sid, srn)
}

type (
	objSet set.Set[ObjectName]
	relSet set.Set[RelationRef]
)

func (m *Model) Reader() (io.Reader, error) {
	b := bytes.Buffer{}
	enc := json.NewEncoder(&b)

	if err := enc.Encode(m); err != nil {
		return nil, err
	}

	return bytes.NewReader(b.Bytes()), nil
}

func (m *Model) Write(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(m)
}

type validationOptions struct {
	allowPermissionInArrowBase bool
}

type ValidationOption func(*validationOptions)

func AllowPermissionInArrowBase(opts *validationOptions) {
	opts.allowPermissionInArrowBase = true
}

// Validate enforces the model's internal consistency.
//
// It enforces the following rules:
//   - Within an object, a permission cannot share the same name as a relation.
//   - Direct relations must reference existing objects .
//   - Wildcard relations must reference existing objects.
//   - Subject relations must reference existing object#relation pairs.
//   - Arrow permissions (relation->rel_or_perm) must reference existing relations/permissions.
func (m *Model) Validate(opts ...ValidationOption) error {
	vOpts := validationOptions{}
	for _, opt := range opts {
		opt(&vOpts)
	}

	validator := newValidator(m, &vOpts)

	return validator.validate()
}

func (m *Model) ValidateRelation(on ObjectName, oid ObjectID, rn RelationName, sn ObjectName, sid ObjectID, srn RelationName) error {
	rel := &relation{on, oid, rn, sn, sid, srn}

	if oid.IsWildcard() {
		return derr.ErrInvalidRelation.Msgf("[%s] object id cannot be a wildcard", rel)
	}

	o := m.Objects[on]
	if o == nil {
		return derr.ErrInvalidRelation.Err(derr.ErrObjectTypeNotFound.Msgf("%s", on)).Msgf("[%s]", rel)
	}

	r := o.Relations[rn]
	if r == nil {
		return derr.ErrInvalidRelation.Err(derr.ErrRelationTypeNotFound.Msgf("%s:%s", on, rn)).Msgf("[%s]", rel)
	}

	// Find all valid assignments for the given subject type.
	refs := lo.Filter(r.Union, func(rr *RelationRef, _ int) bool {
		return rr.Object == rel.sn
	})

	if len(refs) == 0 {
		return derr.ErrInvalidRelation.Msgf("[%s] subject type '%s' is not valid for relation '%s:%s'", rel, rel.sn, on, rn)
	}

	assignment := RelationRef{Object: sn, Relation: srn}

	if rel.sid.IsWildcard() {
		// Wildcard assignment.
		assignment.Relation = WildcardSymbol

		if rel.srn != "" {
			return derr.ErrInvalidRelation.Msgf("[%s] wildcard assignment cannot include subject relation", rel)
		}

		if !lo.ContainsBy(refs, func(rr *RelationRef) bool { return rr.IsWildcard() }) {
			return derr.ErrInvalidRelation.Msgf(
				"[%s] wildcard assignment of '%s' are not allowed on relation '%s:%s'",
				rel, sn, on, rn,
			)
		}
	}

	if !lo.ContainsBy(refs, func(rr *RelationRef) bool { return *rr == assignment }) {
		return derr.ErrInvalidRelation.Msgf("[%s] invalid assignment", rel)
	}

	return nil
}

func (m *Model) StepRelation(r *Relation, subjs ...ObjectName) []*RelationRef {
	steps := lo.FilterMap(r.Union, func(rr *RelationRef, _ int) (*RelationRef, bool) {
		if rr.IsDirect() || rr.IsWildcard() {
			// include direct or wildcard with the expected types.
			return rr, len(subjs) == 0 || lo.Contains(subjs, rr.Object)
		}

		// include subject relations that match or can resolve to the expected types.
		include := len(subjs) == 0 ||
			len(lo.Intersect(m.Objects[rr.Object].Relations[rr.Relation].SubjectTypes, subjs)) > 0 ||
			lo.Contains(subjs, rr.Object)

		return rr, include
	})

	sort.Slice(steps, func(i, j int) bool {
		switch {
		case steps[i].Assignment() > steps[j].Assignment():
			// Wildcard < Subject < Direct
			return true
		case steps[i].Assignment() == steps[j].Assignment():
			return steps[i].String() < steps[j].String()
		default:
			return false
		}
	})

	return steps
}
