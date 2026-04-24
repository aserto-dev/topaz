package model

import (
	"fmt"
	"os"

	set "github.com/deckarep/golang-set/v2"
	"github.com/samber/lo"
)

const (
	ObjectNameSeparator       = "^"
	SubjectRelationSeparator  = "#"
	GeneratedPermissionPrefix = "$"
)

func (m *Model) Invert() *Model {
	if m.inverted == nil {
		fmt.Fprintln(os.Stderr, "load inverted model")

		m.inverted = newInverter(m).invert()
	}

	return m.inverted
}

type inverter struct {
	m     *Model
	im    *Model
	subst map[RelationName]RelationName
}

func newInverter(m *Model) *inverter {
	return &inverter{
		m: m,
		im: &Model{
			Version:  m.Version,
			Objects:  lo.MapValues(m.Objects, func(o *Object, _ ObjectName) *Object { return NewObject() }),
			Metadata: m.Metadata,
		},
		subst: map[RelationName]RelationName{},
	}
}

func (i *inverter) invert() *Model {
	// invert all relations before inverting permissions.
	// this is necessary to create synthetic permissions for subject relations.
	// these permissions are stored in the substitution map (i.subst) and used in inverted permissions.
	for on, o := range i.m.Objects {
		for rn, r := range o.Relations {
			i.invertRelation(on, rn, r)
		}
	}

	for _, o := range i.im.Objects {
		i.applySubstitutions(o)
	}

	for on, o := range i.m.Objects {
		for pn, p := range o.Permissions {
			i.invertPermission(on, pn, o, p)
		}
	}

	for _, o := range i.im.Objects {
		for _, p := range o.Permissions {
			if !p.IsExclusion() {
				continue
			}

			if p.Exclusion.Exclude == nil {
				// It is possible for the 'Exclude' term to be empty in inverted model if the object type
				// cannot have the relation/permission being excluded.
				// In this case, the exclusion permission becomes a single-term union.
				p.Union = PermissionTerms{p.Exclusion.Include}
				p.Exclusion = nil
			}
		}
	}

	return i.im
}

func (i *inverter) invertRelation(on ObjectName, rn RelationName, r *Relation) {
	unionObjs := lo.Associate(r.Union, func(rr *RelationRef) (ObjectName, bool) { return rr.Object, true })

	for _, rr := range r.Union {
		irn := InverseRelation(on, rn, rr.Relation)
		i.addInvertedRelation(on, rr.Object, irn)

		if rr.IsSubject() {
			// add a synthetic permission to reverse the expansion of the subject relation
			srel := i.m.Objects[rr.Object].Relations[rr.Relation]

			for _, subj := range srel.AllRefs() {
				ipr := InverseRelation(on, rn, subj.Relation)
				ipn := PermForRel(ipr)
				p := permissionOrNew(i.im.Objects[subj.Object], ipn, permissionKindUnion)
				i.addSubstitution(ipr, ipn)

				if _, ok := unionObjs[subj.Object]; ok {
					if i.im.Objects[subj.Object].HasRelOrPerm(ipr) {
						p.AddTerm(&PermissionTerm{RelOrPerm: ipr})
					}
				}

				p.AddTerm(&PermissionTerm{Base: InverseRelation(rr.Object, rr.Relation, subj.Relation), RelOrPerm: irn})
			}
		}
	}
}

func (i *inverter) addInvertedRelation(on, ion ObjectName, irn RelationName) {
	o := i.im.Objects[ion]
	o.Relations[irn] = &Relation{Union: []*RelationRef{{Object: on}}}
	// Add a sythetic permission that resolves to the inverted relation.
	// Having these permissions makes it easier to invert permissions that reference
	// subject relations.
	ipn := PermForRel(irn)
	p := permissionOrNew(o, ipn, permissionKindUnion)
	i.addSubstitution(irn, ipn)
	p.AddTerm(&PermissionTerm{RelOrPerm: irn})
}

func (i *inverter) applySubstitutions(o *Object) {
	for pn, p := range o.Permissions {
		for _, pt := range p.Terms() {
			if !pt.IsArrow() && PermForRel(pt.RelOrPerm) == pn {
				continue
			}

			pt.Base = i.sub(pt.Base)
			pt.RelOrPerm = i.sub(pt.RelOrPerm)
		}
	}
}

func (i *inverter) invertPermission(on ObjectName, pn RelationName, o *Object, p *Permission) {
	var typeSet relSet

	switch {
	case p.IsUnion():
		typeSet = set.NewSet(p.Types()...)
	case p.IsIntersection():
		typeSet = lo.Reduce(p.Terms(), func(acc relSet, pt *PermissionTerm, i int) relSet {
			s := set.NewSet(pt.Types()...)
			if i == 0 {
				return s
			}

			if s.IsEmpty() {
				return acc
			}

			return acc.Intersect(s)
		}, nil)
	case p.IsExclusion():
		typeSet = set.NewSet(p.Exclusion.Include.Types()...)
	}

	for _, pt := range p.Terms() {
		newTermInverter(i, on, pn, o, p, pt, typeSet).invert()
	}
}

func (i *inverter) irelSub(on ObjectName, rn RelationName, srn ...RelationName) RelationName {
	return i.sub(InverseRelation(on, rn, srn...))
}

func (i *inverter) sub(rn RelationName) RelationName {
	if pn, ok := i.subst[rn]; ok {
		return pn
	}

	return rn
}

func (i *inverter) addSubstitution(rn, pn RelationName) {
	i.subst[rn] = pn
}

type termInverter struct {
	inv      *inverter
	objName  ObjectName
	permName RelationName
	obj      *Object
	perm     *Permission
	term     *PermissionTerm
	typeSet  relSet
	kind     permissionKind
}

func newTermInverter(i *inverter, on ObjectName, pn RelationName, o *Object, p *Permission, pt *PermissionTerm, typeSet relSet) *termInverter {
	return &termInverter{
		inv:      i,
		objName:  on,
		permName: pn,
		obj:      o,
		perm:     p,
		term:     pt,
		typeSet:  typeSet,
		kind:     kindOf(p),
	}
}

func (ti *termInverter) invert() {
	switch {
	case ti.term.IsArrow():
		// create a subject relation to expand the recursive permission
		ti.invertArrow()

	case ti.obj.HasRelation(ti.term.RelOrPerm):
		ti.invertRelation()

	case ti.obj.HasPermission(ti.term.RelOrPerm):
		ti.invertPermission()
	}
}

func (ti *termInverter) invertArrow() {
	for _, subj := range ti.term.Types() {
		o := ti.inv.im.Objects[subj.Object]
		iName := ti.inv.irelSub(ti.objName, ti.permName, subj.Relation)

		// relation at the arrow's base.
		baseRel := ti.obj.Relations[ti.term.Base]
		for _, baseType := range baseRel.SubjectTypes {
			rp := relOrPerm(ti.inv.m.Objects[baseType], ti.term.RelOrPerm)
			if !rp.TypesContain(subj) {
				continue
			}

			iPerm := permissionOrNew(o, iName, ti.kind)

			base := ti.inv.irelSub(baseType, ti.term.RelOrPerm, subj.Relation)
			tip := ti.inv.irelSub(ti.objName, ti.term.Base)

			iPerm.AddTerm(&PermissionTerm{RelOrPerm: tip, Base: base})
		}
	}
}

func (ti *termInverter) invertRelation() {
	typeRefs := set.NewSet(ti.obj.Relations[ti.term.RelOrPerm].Types()...)
	typeRefs.Intersect(ti.typeSet).Each(func(rr RelationRef) bool {
		iName := InverseRelation(ti.objName, ti.permName, rr.Relation)
		io := ti.inv.im.Objects[rr.Object]

		relOrPerm := ti.inv.irelSub(ti.objName, ti.term.RelOrPerm, rr.Relation)
		if !io.HasRelOrPerm(relOrPerm) {
			// If the relation/permission doesn't exist it may only be defined
			// for wildcards.
			relOrPerm = ti.inv.irelSub(ti.objName, ti.term.RelOrPerm, WildcardSymbol)
		}

		if io.HasRelOrPerm(relOrPerm) {
			// Only add the term if the object type has the relation/permission.
			ip := permissionOrNew(io, iName, ti.kind)
			ip.AddTerm(&PermissionTerm{RelOrPerm: relOrPerm})
		}

		return false // resume iteration
	})
}

func (ti *termInverter) invertPermission() {
	typeRefs := set.NewSet(types(ti.obj, ti.term.RelOrPerm)...)
	typeRefs.Intersect(ti.typeSet).Each(func(rr RelationRef) bool {
		iName := InverseRelation(ti.objName, ti.permName, rr.Relation)
		ip := permissionOrNew(ti.inv.im.Objects[rr.Object], iName, ti.kind)
		ip.AddTerm(&PermissionTerm{RelOrPerm: ti.inv.irelSub(ti.objName, ti.term.RelOrPerm, rr.Relation)})

		return false // resume iteration
	})
}

type RelOrPerm interface {
	TypesContain(rel RelationRef) bool
}

func relOrPerm(o *Object, rn RelationName) RelOrPerm { //nolint:ireturn  // abstraction over relations and permissions
	if o.HasRelation(rn) {
		return o.Relations[rn]
	}

	return o.Permissions[rn]
}

type permissionKind int

const (
	permissionKindUnion permissionKind = iota
	permissionKindIntersection
	permissionKindExclusion
)

func kindOf(p *Permission) permissionKind {
	switch {
	case p.IsUnion():
		return permissionKindUnion
	case p.IsIntersection():
		return permissionKindIntersection
	case p.IsExclusion():
		return permissionKindExclusion
	}

	panic("unknown permission kind")
}

func permissionOrNew(o *Object, pn RelationName, kind permissionKind) *Permission {
	p := o.Permissions[pn]
	if p != nil {
		return p
	}

	p = &Permission{}
	terms := PermissionTerms{}

	switch kind {
	case permissionKindUnion:
		p.Union = terms
	case permissionKindIntersection:
		p.Intersection = terms
	case permissionKindExclusion:
		p.Exclusion = &ExclusionPermission{}
	}

	o.Permissions[pn] = p

	return p
}

func InverseRelation(on ObjectName, rn RelationName, srn ...RelationName) RelationName {
	irn := RelationName(fmt.Sprintf("%s%s%s", on, ObjectNameSeparator, rn))

	switch {
	case len(srn) == 0 || srn[0] == "":
		return irn
	case len(srn) == 1:
		return RelationName(fmt.Sprintf("%s%s%s", irn, SubjectRelationSeparator, srn[0]))
	default:
		panic("too many subject relations")
	}
}

func PermForRel(rn RelationName) RelationName {
	return RelationName(fmt.Sprintf("%s%s", GeneratedPermissionPrefix, rn))
}

func types(o *Object, rn RelationName) RelationRefs {
	if o.HasRelation(rn) {
		return o.Relations[rn].Types()
	}

	return o.Permissions[rn].Types()
}
