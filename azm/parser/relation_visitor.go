package parser

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/aserto-dev/topaz/azm/model"
	"github.com/samber/lo"
)

type RelationVisitor struct {
	BaseAzmVisitor
}

func (v *RelationVisitor) Visit(tree antlr.ParseTree) any {
	switch t := tree.(type) {
	case *RelationContext:
		return t.Accept(v)
	default:
		panic("RelationVisitor can only visit relations")
	}
}

func (v *RelationVisitor) VisitRelation(c *RelationContext) any {
	return lo.Map(c.AllRel(), func(rel IRelContext, _ int) *model.RelationRef {
		if term, ok := rel.Accept(v).(*model.RelationRef); ok {
			return term
		}

		return &model.RelationRef{}
	})
}

func (v *RelationVisitor) VisitDirectRel(c *DirectRelContext) any {
	return &model.RelationRef{Object: model.ObjectName(c.ID().GetText())}
}

func (v *RelationVisitor) VisitWildcardRel(c *WildcardRelContext) any {
	return &model.RelationRef{
		Object:   model.ObjectName(c.ID().GetText()),
		Relation: model.WildcardSymbol,
	}
}

func (v *RelationVisitor) VisitSubjectRel(c *SubjectRelContext) any {
	return &model.RelationRef{
		Object:   model.ObjectName(c.ID(0).GetText()),
		Relation: model.RelationName(c.ID(1).GetText()),
	}
}
