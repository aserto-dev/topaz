package model

import (
	"fmt"

	"github.com/aserto-dev/topaz/api/directory/pkg/derr"

	set "github.com/deckarep/golang-set/v2"
	"github.com/hashicorp/go-multierror"
	"github.com/samber/lo"
)

type termRef struct {
	perm *RelationRef
	term *PermissionTerm
}

type validator struct {
	*Model

	opts     *validationOptions
	deferred []termRef
}

func newValidator(m *Model, opts *validationOptions) *validator {
	return &validator{Model: m, opts: opts}
}

func (v *validator) validate() error {
	// Pass 1 (syntax): ensure no name collisions and all relations reference existing objects/relations.
	if err := v.validateReferences(); err != nil {
		return derr.ErrInvalidArgument.Err(err)
	}

	// Pass 2: resolve all relations to a set of possible subject types.
	if err := v.resolveRelations(); err != nil {
		return derr.ErrInvalidArgument.Err(err)
	}

	// Pass 3: validate all arrow operators in permissions. This requires that all relations have already been resolved.
	if err := v.validatePermissions(); err != nil {
		return derr.ErrInvalidArgument.Err(err)
	}

	// Pass 4: resolve all permissions to a set of possible subject types.
	if err := v.resolvePermissions(); err != nil {
		return derr.ErrInvalidArgument.Err(err)
	}

	return nil
}

func (v *validator) validateReferences() error {
	var errs error

	for on, o := range v.Objects {
		if err := v.validateObjectRels(on, o); err != nil {
			errs = multierror.Append(errs, err)
		}

		if err := v.validateObjectPerms(on, o); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs
}

func (v *validator) validateObjectRels(on ObjectName, o *Object) error {
	var errs error

	for rn, rs := range o.Relations {
		for _, r := range rs.Union {
			o := v.Objects[r.Object]
			if o == nil {
				errs = multierror.Append(errs, derr.ErrInvalidRelationType.Msgf(
					"relation '%s:%s' references undefined object type '%s'", on, rn, r.Object),
				)

				continue
			}

			if r.IsSubject() {
				if _, ok := o.Relations[r.Relation]; !ok {
					errs = multierror.Append(errs, derr.ErrInvalidRelationType.Msgf(
						"relation '%s:%s' references undefined relation type '%s#%s'", on, rn, r.Object, r.Relation),
					)
				}
			}
		}
	}

	return errs
}

func (v *validator) validateObjectPerms(on ObjectName, o *Object) error {
	var errs error

	for pn, p := range o.Permissions {
		terms := p.Terms()
		if len(terms) == 0 {
			errs = multierror.Append(errs, derr.ErrInvalidPermission.Msgf(
				"permission '%s:%s' has no definition", on, pn),
			)

			continue
		}

		for _, term := range terms {
			if term == nil {
				errs = multierror.Append(errs, derr.ErrInvalidPermission.Msgf(
					"permission '%s:%s' has an empty term", on, pn),
				)

				continue
			}

			switch {
			case term.IsArrow():
				// this is an arrow operator.
				// validate that the base relation exists on this object type.
				// at this stage we don't yet resolve the relation to a set of subject types.
				if !o.HasRelOrPerm(term.Base) {
					errs = multierror.Append(errs, derr.ErrInvalidPermission.Msgf(
						"permission '%s:%s' references undefined relation type '%s:%s'", on, pn, on, term.Base),
					)
				}

			default:
				// validate that the relation/permission exists on this object type.
				if !o.HasRelOrPerm(term.RelOrPerm) {
					errs = multierror.Append(errs, derr.ErrInvalidPermission.Msgf(
						"permission '%s:%s' references undefined relation or permission '%s:%s'", on, pn, on, term.RelOrPerm),
					)
				}
			}
		}
	}

	return errs
}

func (v *validator) validatePermissions() error {
	var errs error

	for on, o := range v.Objects {
		for pn, p := range o.Permissions {
			if err := v.validatePermission(on, pn, p); err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}

	return errs
}

func (v *validator) validatePermission(on ObjectName, pn RelationName, p *Permission) error {
	o := v.Objects[on]

	var errs error

	for _, term := range p.Terms() {
		if !term.IsArrow() {
			continue
		}
		// given a reference base->rel_or_perm, validate that all object types that `base` can resolve to
		// have a permission or relation named `rel_or_perm`.
		if o.HasPermission(term.Base) {
			if !v.opts.allowPermissionInArrowBase {
				errs = multierror.Append(errs, derr.ErrInvalidPermission.Msgf(
					"permission '%s:%s' references permission '%s', which is not allowed in arrow base. only relations can be used.",
					on, pn, term.Base,
				))
			}

			continue
		}

		r := o.Relations[term.Base]

		for _, ref := range r.Union {
			if ref.IsWildcard() {
				errs = multierror.Append(errs, derr.ErrInvalidPermission.Msgf(
					"wildcard relation '%s:%s' not allowed in the base of an arrow operator '%s%s%s' in permission '%s:%s'",
					on, term.Base, term.Base, ArrowSymbol, term.RelOrPerm, on, pn,
				))
			}
		}

		for _, st := range r.SubjectTypes {
			if !v.Objects[st].HasRelOrPerm(term.RelOrPerm) {
				arrow := fmt.Sprintf("%s%s%s", term.Base, ArrowSymbol, term.RelOrPerm)
				errs = multierror.Append(errs, derr.ErrInvalidPermission.Msgf(
					"permission '%s:%s' references '%s', which can resolve to undefined relation or permission '%s:%s' ",
					on, pn, arrow, st, term.RelOrPerm,
				))
			}
		}
	}

	return errs
}

func (v *validator) resolveRelations() error {
	var errs error

	for on, o := range v.Objects {
		for rn, r := range o.Relations {
			seen := set.NewSet(RelationRef{Object: on, Relation: rn})
			subs, intermediates := v.resolveRelation(r, seen)

			switch len(subs) {
			case 0:
				errs = multierror.Append(errs, derr.ErrInvalidRelationType.Msgf(
					"relation '%s:%s' is circular and does not resolve to any object types", on, rn),
				)
			default:
				r.SubjectTypes = subs
				r.Intermediates = intermediates
			}
		}
	}

	return errs
}

func (v *validator) resolveRelation(r *Relation, seen relSet) ([]ObjectName, RelationRefs) {
	if len(r.SubjectTypes) > 0 {
		// already resolved
		return r.SubjectTypes, r.Intermediates
	}

	subjectTypes := set.NewSet[ObjectName]()
	intermediateTypes := set.NewSet[RelationRef]()

	for _, rr := range r.Union {
		if rr.IsSubject() {
			intermediateTypes.Add(*rr)

			if !seen.Contains(*rr) {
				seen.Add(*rr)
				subs, intermediates := v.resolveRelation(v.Objects[rr.Object].Relations[rr.Relation], seen)
				subjectTypes.Append(subs...)
				intermediateTypes.Append(intermediates...)
			}
		} else {
			subjectTypes.Add(rr.Object)
		}
	}

	return subjectTypes.ToSlice(), intermediateTypes.ToSlice()
}

func (v *validator) resolvePermissions() error {
	seen := set.NewSet[RelationRef]()

	for on, o := range v.Objects {
		for pn := range o.Permissions {
			v.resolvePermission(&RelationRef{on, pn}, seen)
		}
	}

	// resolve subject types of cyclic permission terms.
	v.resolveCyclicTerms()

	var errs error

	for on, o := range v.Objects {
		for pn, p := range o.Permissions {
			if len(p.SubjectTypes) == 0 {
				errs = multierror.Append(errs, derr.ErrInvalidPermission.Msgf(
					"permission '%s:%s' cannot be satisfied by any type", on, pn),
				)
			}
		}
	}

	return errs
}

func (v *validator) resolvePermission(ref *RelationRef, seen relSet) (objSet, relSet) {
	p, ok := v.Objects[ref.Object].Permissions[ref.Relation]
	if !ok {
		// No such permission. Most likely a bug in the model inversion logic.
		// Return empty sets which result in a validation error.
		return set.NewSet[ObjectName](), set.NewSet[RelationRef]()
	}

	if len(p.SubjectTypes) > 0 {
		// already resolved
		return set.NewSet(p.SubjectTypes...), set.NewSet(p.Intermediates...)
	}

	if seen.Contains(*ref) {
		// cycle detected
		return set.NewSet[ObjectName](), set.NewSet[RelationRef]()
	}

	seen.Add(*ref)

	for _, term := range p.Terms() {
		term.SubjectTypes, term.Intermediates = v.resolvePermissionTerm(termRef{ref, term}, seen)
	}

	// filter out terms that have no subject types. They represent cycles that are still being resolved.
	resolvedTerms := lo.Filter(p.Terms(), func(term *PermissionTerm, _ int) bool {
		return len(term.SubjectTypes) > 0
	})

	var (
		subjTypes     objSet
		intermediates relSet
	)

	switch {
	case p.IsUnion():
		subjTypes = set.NewSet(lo.FlatMap(resolvedTerms, func(term *PermissionTerm, _ int) []ObjectName {
			return term.SubjectTypes
		})...)
		intermediates = set.NewSet(lo.FlatMap(resolvedTerms, func(term *PermissionTerm, _ int) []RelationRef {
			return term.Intermediates
		})...)

	case p.IsIntersection():
		subjTypes = lo.Reduce(resolvedTerms, func(acc objSet, term *PermissionTerm, i int) objSet {
			subjs := set.NewSet(term.SubjectTypes...)

			if i == 0 {
				return subjs
			}

			return acc.Intersect(subjs)
		}, nil)
		intermediates = lo.Reduce(resolvedTerms, func(acc relSet, term *PermissionTerm, i int) relSet {
			subjs := set.NewSet(term.Intermediates...)

			if i == 0 {
				return subjs
			}

			return acc.Intersect(subjs)
		}, nil)

	case p.IsExclusion():
		subjTypes = set.NewSet(p.Exclusion.Include.SubjectTypes...)
		intermediates = set.NewSet(p.Exclusion.Include.Intermediates...)
	}

	p.SubjectTypes = subjTypes.ToSlice()
	p.Intermediates = intermediates.ToSlice()

	return subjTypes, intermediates
}

func (v *validator) resolvePermissionTerm(t termRef, seen relSet) ([]ObjectName, RelationRefs) {
	term, perm := t.term, t.perm
	base, tip := term.Base, term.RelOrPerm

	var baseRefs set.Set[RelationRef]

	intermediates := set.NewSet[RelationRef]()

	switch {
	case term.IsArrow():
		var sts []ObjectName

		o := v.Objects[perm.Object]
		if o.HasRelation(base) {
			sts = o.Relations[base].SubjectTypes
			intermediates.Append(o.Relations[base].Intermediates...)
		} else {
			types, interims := v.resolvePermission(&RelationRef{Object: perm.Object, Relation: base}, seen)
			sts = types.ToSlice()
			intermediates.Append(interims.ToSlice()...)
		}

		baseRefs = set.NewSet(lo.Map(sts, func(st ObjectName, _ int) RelationRef {
			return RelationRef{Object: st, Relation: tip}
		})...)

	default:
		baseRefs = set.NewSet(RelationRef{Object: perm.Object, Relation: tip})
	}

	subjectTypes := set.NewSet[ObjectName]()

	for baseRef := range baseRefs.Iter() {
		o := v.Objects[baseRef.Object]

		if o.HasRelation(baseRef.Relation) {
			// Relations are already resolved to a set of subject types.
			subjectTypes.Append(o.Relations[baseRef.Relation].SubjectTypes...)
			intermediates.Append(o.Relations[baseRef.Relation].Intermediates...)

			continue
		}

		sts, interims := v.resolvePermission(&baseRef, seen)
		if sts.IsEmpty() {
			v.deferred = append(v.deferred, t)
		}

		intermediates.Append(interims.ToSlice()...)

		subjectTypes = subjectTypes.Union(sts)
	}

	return subjectTypes.ToSlice(), intermediates.ToSlice()
}

// resolvedCyclicTerms fills in the set of subject and intermediate types on
// self-referential permission terms (i.e. type cycles) that can only be
// resolved once all other terms have been resolved.
func (v *validator) resolveCyclicTerms() {
	for _, t := range v.deferred {
		perm, term := t.perm, t.term
		base, tip := term.Base, term.RelOrPerm

		var baseRefs set.Set[RelationRef]

		intermediates := set.NewSet[RelationRef]()

		if term.IsArrow() {
			var subjTypes []ObjectName

			o := v.Objects[perm.Object]
			if o.HasRelation(base) {
				rel := o.Relations[base]
				subjTypes = rel.SubjectTypes
				intermediates.Append(rel.Intermediates...)
			} else {
				perm := o.Permissions[base]
				subjTypes = perm.SubjectTypes
				intermediates.Append(perm.Intermediates...)
			}

			baseRefs = set.NewSet(
				lo.Map(subjTypes, func(on ObjectName, _ int) RelationRef {
					return RelationRef{Object: on, Relation: tip}
				})...,
			)
		} else {
			baseRefs = set.NewSet(RelationRef{Object: perm.Object, Relation: tip})
		}

		subjectTypes := set.NewSet[ObjectName]()

		for baseRef := range baseRefs.Iter() {
			o := v.Objects[baseRef.Object]
			if o.HasRelation(baseRef.Relation) {
				subjectTypes.Append(o.Relations[baseRef.Relation].SubjectTypes...)
				intermediates.Append(o.Relations[baseRef.Relation].Intermediates...)
			} else {
				subjectTypes.Append(o.Permissions[baseRef.Relation].SubjectTypes...)
				intermediates.Append(o.Permissions[baseRef.Relation].Intermediates...)
			}
		}

		term.SubjectTypes, term.Intermediates = subjectTypes.ToSlice(), intermediates.ToSlice()
	}
}
