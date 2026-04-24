package model

import (
	"fmt"

	"github.com/aserto-dev/topaz/azm/internal/lox"
	"github.com/samber/lo"
)

type (
	ObjectName   string
	RelationName string
)

func (on ObjectName) String() string {
	return string(on)
}

func (rn RelationName) String() string {
	return string(rn)
}

type (
	Relations   map[RelationName]*Relation
	Permissions map[RelationName]*Permission
)

type Object struct {
	Relations   Relations   `json:"relations,omitempty"`
	Permissions Permissions `json:"permissions,omitempty"`
}

func NewObject() *Object {
	return &Object{
		Relations:   Relations{},
		Permissions: Permissions{},
	}
}

func (o *Object) HasRelation(name RelationName) bool {
	if o == nil {
		return false
	}

	return o.Relations[name] != nil
}

func (o *Object) HasPermission(name RelationName) bool {
	if o == nil {
		return false
	}

	return o.Permissions[name] != nil
}

func (o *Object) HasRelOrPerm(name RelationName) bool {
	return o.HasRelation(name) || o.HasPermission(name)
}

// SubjectTypes returns the list of possible subject types for the given relation or permission.
func (o *Object) SubjectTypes(name RelationName) []ObjectName {
	if o == nil {
		return nil
	}

	if r := o.Relations[name]; r != nil {
		return r.SubjectTypes
	}

	if p := o.Permissions[name]; p != nil {
		return p.SubjectTypes
	}

	return nil
}

type Relation struct {
	Union         []*RelationRef `json:"union,omitempty"`
	SubjectTypes  []ObjectName   `json:"subject_types,omitempty"`
	Intermediates RelationRefs   `json:"intermediates,omitempty"`
}

func (r *Relation) Types() RelationRefs {
	return append(
		lo.Map(r.SubjectTypes, func(on ObjectName, _ int) RelationRef {
			return RelationRef{Object: on}
		}),
		r.Intermediates...,
	)
}

func (r *Relation) AllRefs() []RelationRef {
	return append(lo.Map(r.SubjectTypes, func(on ObjectName, _ int) RelationRef {
		return RelationRef{Object: on}
	}), r.Intermediates...)
}

func (r *Relation) AddRef(rr *RelationRef) {
	if !lox.ContainsPtr(r.Union, rr) {
		r.Union = append(r.Union, rr)
	}
}

func (r *Relation) TypesContain(rr RelationRef) bool {
	if rr.IsSubject() {
		return lo.Contains(r.Intermediates, rr)
	}

	return lo.Contains(r.SubjectTypes, rr.Object)
}

type RelationRefs []RelationRef

type RelationRef struct {
	Object   ObjectName   `json:"object,omitempty"`
	Relation RelationName `json:"relation,omitempty"`
}

type RelationAssignment int

const (
	RelationAssignmentUnknown RelationAssignment = iota
	RelationAssignmentDirect
	RelationAssignmentSubject
	RelationAssignmentWildcard
)

func NewRelationRef(on ObjectName, rn RelationName) *RelationRef {
	return &RelationRef{Object: on, Relation: rn}
}

func (rr *RelationRef) String() string {
	switch {
	case rr.IsWildcard():
		return fmt.Sprintf("%s:%s", rr.Object, rr.Relation)
	case rr.IsDirect():
		return string(rr.Object)
	case rr.IsSubject():
		return fmt.Sprintf("%s#%s", rr.Object, rr.Relation)
	}

	panic("unknown relation assignment")
}

func (rr *RelationRef) Assignment() RelationAssignment {
	if rr == nil {
		return RelationAssignmentUnknown
	}

	switch {
	case rr.Relation == WildcardSymbol:
		return RelationAssignmentWildcard
	case rr.Relation != "":
		return RelationAssignmentSubject
	case rr.Object != "":
		return RelationAssignmentDirect
	default:
		return RelationAssignmentUnknown
	}
}

func (rr *RelationRef) IsDirect() bool {
	return rr.Assignment() == RelationAssignmentDirect
}

func (rr *RelationRef) IsWildcard() bool {
	return rr.Assignment() == RelationAssignmentWildcard
}

func (rr *RelationRef) IsSubject() bool {
	return rr.Assignment() == RelationAssignmentSubject
}

type Permission struct {
	Union        PermissionTerms      `json:"union,omitempty"`
	Intersection PermissionTerms      `json:"intersection,omitempty"`
	Exclusion    *ExclusionPermission `json:"exclusion,omitempty"`

	SubjectTypes  []ObjectName `json:"subject_types,omitempty"`
	Intermediates RelationRefs `json:"intermediates,omitempty"`
}

func (p *Permission) IsUnion() bool {
	return p.Union != nil
}

func (p *Permission) IsIntersection() bool {
	return p.Intersection != nil
}

func (p *Permission) IsExclusion() bool {
	return p.Exclusion != nil
}

func (p *Permission) Terms() []*PermissionTerm {
	var refs []*PermissionTerm

	switch {
	case p.IsUnion():
		refs = p.Union
	case p.IsIntersection():
		refs = p.Intersection
	case p.IsExclusion():
		refs = []*PermissionTerm{p.Exclusion.Include, p.Exclusion.Exclude}
	}

	return refs
}

func (p *Permission) AddTerm(pt *PermissionTerm) {
	switch {
	case p.IsUnion() && !p.Union.Contains(pt):
		p.Union = append(p.Union, pt)
	case p.IsIntersection() && !p.Intersection.Contains(pt):
		p.Intersection = append(p.Intersection, pt)
	case p.IsExclusion():
		if p.Exclusion.Include == nil {
			p.Exclusion.Include = pt
		} else {
			p.Exclusion.Exclude = pt
		}
	}
}

func (p *Permission) Types() RelationRefs {
	return append(objectNamesToRelationRefs(p.SubjectTypes), p.Intermediates...)
}

func (p *Permission) TypesContain(rr RelationRef) bool {
	if rr.IsSubject() {
		return lo.Contains(p.Intermediates, rr)
	}

	return lo.Contains(p.SubjectTypes, rr.Object)
}

type PermissionTerm struct {
	Base      RelationName `json:"base,omitempty"`
	RelOrPerm RelationName `json:"rel_or_perm"`

	SubjectTypes  []ObjectName `json:"subject_types,omitempty"`
	Intermediates RelationRefs `json:"intermediates,omitempty"`
}

func (pr *PermissionTerm) String() string {
	if pr == nil {
		return "<nil>"
	}

	s := string(pr.RelOrPerm)

	if pr.Base != "" {
		s = string(pr.Base) + ArrowSymbol + s
	}

	return s
}

func (pr *PermissionTerm) IsArrow() bool {
	return pr.Base != ""
}

func (pr *PermissionTerm) Types() RelationRefs {
	return append(
		objectNamesToRelationRefs(pr.SubjectTypes),
		pr.Intermediates...,
	)
}

type PermissionTerms []*PermissionTerm

func (pts PermissionTerms) Contains(pt *PermissionTerm) bool {
	for _, t := range pts {
		if t.Base == pt.Base && t.RelOrPerm == pt.RelOrPerm {
			return true
		}
	}

	return false
}

func objectNamesToRelationRefs(names []ObjectName) RelationRefs {
	return lo.Map(names, func(on ObjectName, _ int) RelationRef {
		return RelationRef{Object: on}
	})
}

type ExclusionPermission struct {
	Include *PermissionTerm `json:"include,omitempty"`
	Exclude *PermissionTerm `json:"exclude,omitempty"`
}

type ArrowPermission struct {
	Relation   string `json:"relation,omitempty"`
	Permission string `json:"permission,omitempty"`
}
