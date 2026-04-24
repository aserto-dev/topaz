// Code generated from Azm.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // Azm
import "github.com/antlr4-go/antlr/v4"

type BaseAzmVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseAzmVisitor) VisitRelation(ctx *RelationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseAzmVisitor) VisitUnionPerm(ctx *UnionPermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseAzmVisitor) VisitIntersectionPerm(ctx *IntersectionPermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseAzmVisitor) VisitExclusionPerm(ctx *ExclusionPermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseAzmVisitor) VisitUnion(ctx *UnionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseAzmVisitor) VisitIntersection(ctx *IntersectionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseAzmVisitor) VisitExclusion(ctx *ExclusionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseAzmVisitor) VisitDirectRel(ctx *DirectRelContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseAzmVisitor) VisitWildcardRel(ctx *WildcardRelContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseAzmVisitor) VisitSubjectRel(ctx *SubjectRelContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseAzmVisitor) VisitDirectPerm(ctx *DirectPermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseAzmVisitor) VisitArrowPerm(ctx *ArrowPermContext) interface{} {
	return v.VisitChildren(ctx)
}
