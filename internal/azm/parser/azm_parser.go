// Code generated from Azm.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // Azm
import (
	"fmt"
	"strconv"
	"sync"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = strconv.Itoa
var _ = sync.Once{}

type AzmParser struct {
	*antlr.BaseParser
}

var AzmParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func azmParserInit() {
	staticData := &AzmParserStaticData
	staticData.LiteralNames = []string{
		"", "'|'", "'&'", "'-'", "'->'", "'#'", "':'", "'*'",
	}
	staticData.SymbolicNames = []string{
		"", "", "", "", "ARROW", "HASH", "COLON", "ASTERISK", "ID", "WS", "ERROR",
	}
	staticData.RuleNames = []string{
		"relation", "permission", "union", "intersection", "exclusion", "rel",
		"perm",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 10, 73, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 1, 0, 1, 0, 1, 0, 5, 0, 18, 8, 0, 10, 0, 12,
		0, 21, 9, 0, 1, 0, 1, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 3, 1, 34, 8, 1, 1, 2, 1, 2, 1, 2, 5, 2, 39, 8, 2, 10, 2, 12, 2, 42,
		9, 2, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 5, 3, 49, 8, 3, 10, 3, 12, 3, 52, 9,
		3, 1, 4, 1, 4, 1, 4, 1, 4, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 3,
		5, 65, 8, 5, 1, 6, 1, 6, 1, 6, 1, 6, 3, 6, 71, 8, 6, 1, 6, 0, 0, 7, 0,
		2, 4, 6, 8, 10, 12, 0, 0, 73, 0, 14, 1, 0, 0, 0, 2, 33, 1, 0, 0, 0, 4,
		35, 1, 0, 0, 0, 6, 43, 1, 0, 0, 0, 8, 53, 1, 0, 0, 0, 10, 64, 1, 0, 0,
		0, 12, 70, 1, 0, 0, 0, 14, 19, 3, 10, 5, 0, 15, 16, 5, 1, 0, 0, 16, 18,
		3, 10, 5, 0, 17, 15, 1, 0, 0, 0, 18, 21, 1, 0, 0, 0, 19, 17, 1, 0, 0, 0,
		19, 20, 1, 0, 0, 0, 20, 22, 1, 0, 0, 0, 21, 19, 1, 0, 0, 0, 22, 23, 5,
		0, 0, 1, 23, 1, 1, 0, 0, 0, 24, 25, 3, 4, 2, 0, 25, 26, 5, 0, 0, 1, 26,
		34, 1, 0, 0, 0, 27, 28, 3, 6, 3, 0, 28, 29, 5, 0, 0, 1, 29, 34, 1, 0, 0,
		0, 30, 31, 3, 8, 4, 0, 31, 32, 5, 0, 0, 1, 32, 34, 1, 0, 0, 0, 33, 24,
		1, 0, 0, 0, 33, 27, 1, 0, 0, 0, 33, 30, 1, 0, 0, 0, 34, 3, 1, 0, 0, 0,
		35, 40, 3, 12, 6, 0, 36, 37, 5, 1, 0, 0, 37, 39, 3, 12, 6, 0, 38, 36, 1,
		0, 0, 0, 39, 42, 1, 0, 0, 0, 40, 38, 1, 0, 0, 0, 40, 41, 1, 0, 0, 0, 41,
		5, 1, 0, 0, 0, 42, 40, 1, 0, 0, 0, 43, 44, 3, 12, 6, 0, 44, 45, 5, 2, 0,
		0, 45, 50, 3, 12, 6, 0, 46, 47, 5, 2, 0, 0, 47, 49, 3, 12, 6, 0, 48, 46,
		1, 0, 0, 0, 49, 52, 1, 0, 0, 0, 50, 48, 1, 0, 0, 0, 50, 51, 1, 0, 0, 0,
		51, 7, 1, 0, 0, 0, 52, 50, 1, 0, 0, 0, 53, 54, 3, 12, 6, 0, 54, 55, 5,
		3, 0, 0, 55, 56, 3, 12, 6, 0, 56, 9, 1, 0, 0, 0, 57, 65, 5, 8, 0, 0, 58,
		59, 5, 8, 0, 0, 59, 60, 5, 6, 0, 0, 60, 65, 5, 7, 0, 0, 61, 62, 5, 8, 0,
		0, 62, 63, 5, 5, 0, 0, 63, 65, 5, 8, 0, 0, 64, 57, 1, 0, 0, 0, 64, 58,
		1, 0, 0, 0, 64, 61, 1, 0, 0, 0, 65, 11, 1, 0, 0, 0, 66, 71, 5, 8, 0, 0,
		67, 68, 5, 8, 0, 0, 68, 69, 5, 4, 0, 0, 69, 71, 5, 8, 0, 0, 70, 66, 1,
		0, 0, 0, 70, 67, 1, 0, 0, 0, 71, 13, 1, 0, 0, 0, 6, 19, 33, 40, 50, 64,
		70,
	}
	deserializer := antlr.NewATNDeserializer(nil)
	staticData.atn = deserializer.Deserialize(staticData.serializedATN)
	atn := staticData.atn
	staticData.decisionToDFA = make([]*antlr.DFA, len(atn.DecisionToState))
	decisionToDFA := staticData.decisionToDFA
	for index, state := range atn.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(state, index)
	}
}

// AzmParserInit initializes any static state used to implement AzmParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewAzmParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func AzmParserInit() {
	staticData := &AzmParserStaticData
	staticData.once.Do(azmParserInit)
}

// NewAzmParser produces a new parser instance for the optional input antlr.TokenStream.
func NewAzmParser(input antlr.TokenStream) *AzmParser {
	AzmParserInit()
	this := new(AzmParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &AzmParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "Azm.g4"

	return this
}

// AzmParser tokens.
const (
	AzmParserEOF      = antlr.TokenEOF
	AzmParserT__0     = 1
	AzmParserT__1     = 2
	AzmParserT__2     = 3
	AzmParserARROW    = 4
	AzmParserHASH     = 5
	AzmParserCOLON    = 6
	AzmParserASTERISK = 7
	AzmParserID       = 8
	AzmParserWS       = 9
	AzmParserERROR    = 10
)

// AzmParser rules.
const (
	AzmParserRULE_relation     = 0
	AzmParserRULE_permission   = 1
	AzmParserRULE_union        = 2
	AzmParserRULE_intersection = 3
	AzmParserRULE_exclusion    = 4
	AzmParserRULE_rel          = 5
	AzmParserRULE_perm         = 6
)

// IRelationContext is an interface to support dynamic dispatch.
type IRelationContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllRel() []IRelContext
	Rel(i int) IRelContext
	EOF() antlr.TerminalNode

	// IsRelationContext differentiates from other interfaces.
	IsRelationContext()
}

type RelationContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyRelationContext() *RelationContext {
	var p = new(RelationContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = AzmParserRULE_relation
	return p
}

func InitEmptyRelationContext(p *RelationContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = AzmParserRULE_relation
}

func (*RelationContext) IsRelationContext() {}

func NewRelationContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *RelationContext {
	var p = new(RelationContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = AzmParserRULE_relation

	return p
}

func (s *RelationContext) GetParser() antlr.Parser { return s.parser }

func (s *RelationContext) AllRel() []IRelContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IRelContext); ok {
			len++
		}
	}

	tst := make([]IRelContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IRelContext); ok {
			tst[i] = t.(IRelContext)
			i++
		}
	}

	return tst
}

func (s *RelationContext) Rel(i int) IRelContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IRelContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IRelContext)
}

func (s *RelationContext) EOF() antlr.TerminalNode {
	return s.GetToken(AzmParserEOF, 0)
}

func (s *RelationContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RelationContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *RelationContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case AzmVisitor:
		return t.VisitRelation(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *AzmParser) Relation() (localctx IRelationContext) {
	localctx = NewRelationContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, AzmParserRULE_relation)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(14)
		p.Rel()
	}
	p.SetState(19)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == AzmParserT__0 {
		{
			p.SetState(15)
			p.Match(AzmParserT__0)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(16)
			p.Rel()
		}

		p.SetState(21)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(22)
		p.Match(AzmParserEOF)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IPermissionContext is an interface to support dynamic dispatch.
type IPermissionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsPermissionContext differentiates from other interfaces.
	IsPermissionContext()
}

type PermissionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPermissionContext() *PermissionContext {
	var p = new(PermissionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = AzmParserRULE_permission
	return p
}

func InitEmptyPermissionContext(p *PermissionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = AzmParserRULE_permission
}

func (*PermissionContext) IsPermissionContext() {}

func NewPermissionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PermissionContext {
	var p = new(PermissionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = AzmParserRULE_permission

	return p
}

func (s *PermissionContext) GetParser() antlr.Parser { return s.parser }

func (s *PermissionContext) CopyAll(ctx *PermissionContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *PermissionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PermissionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type ExclusionPermContext struct {
	PermissionContext
}

func NewExclusionPermContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ExclusionPermContext {
	var p = new(ExclusionPermContext)

	InitEmptyPermissionContext(&p.PermissionContext)
	p.parser = parser
	p.CopyAll(ctx.(*PermissionContext))

	return p
}

func (s *ExclusionPermContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExclusionPermContext) Exclusion() IExclusionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExclusionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExclusionContext)
}

func (s *ExclusionPermContext) EOF() antlr.TerminalNode {
	return s.GetToken(AzmParserEOF, 0)
}

func (s *ExclusionPermContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case AzmVisitor:
		return t.VisitExclusionPerm(s)

	default:
		return t.VisitChildren(s)
	}
}

type IntersectionPermContext struct {
	PermissionContext
}

func NewIntersectionPermContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *IntersectionPermContext {
	var p = new(IntersectionPermContext)

	InitEmptyPermissionContext(&p.PermissionContext)
	p.parser = parser
	p.CopyAll(ctx.(*PermissionContext))

	return p
}

func (s *IntersectionPermContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *IntersectionPermContext) Intersection() IIntersectionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIntersectionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIntersectionContext)
}

func (s *IntersectionPermContext) EOF() antlr.TerminalNode {
	return s.GetToken(AzmParserEOF, 0)
}

func (s *IntersectionPermContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case AzmVisitor:
		return t.VisitIntersectionPerm(s)

	default:
		return t.VisitChildren(s)
	}
}

type UnionPermContext struct {
	PermissionContext
}

func NewUnionPermContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *UnionPermContext {
	var p = new(UnionPermContext)

	InitEmptyPermissionContext(&p.PermissionContext)
	p.parser = parser
	p.CopyAll(ctx.(*PermissionContext))

	return p
}

func (s *UnionPermContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *UnionPermContext) Union() IUnionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IUnionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IUnionContext)
}

func (s *UnionPermContext) EOF() antlr.TerminalNode {
	return s.GetToken(AzmParserEOF, 0)
}

func (s *UnionPermContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case AzmVisitor:
		return t.VisitUnionPerm(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *AzmParser) Permission() (localctx IPermissionContext) {
	localctx = NewPermissionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, AzmParserRULE_permission)
	p.SetState(33)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 1, p.GetParserRuleContext()) {
	case 1:
		localctx = NewUnionPermContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(24)
			p.Union()
		}
		{
			p.SetState(25)
			p.Match(AzmParserEOF)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		localctx = NewIntersectionPermContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(27)
			p.Intersection()
		}
		{
			p.SetState(28)
			p.Match(AzmParserEOF)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		localctx = NewExclusionPermContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(30)
			p.Exclusion()
		}
		{
			p.SetState(31)
			p.Match(AzmParserEOF)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IUnionContext is an interface to support dynamic dispatch.
type IUnionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllPerm() []IPermContext
	Perm(i int) IPermContext

	// IsUnionContext differentiates from other interfaces.
	IsUnionContext()
}

type UnionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyUnionContext() *UnionContext {
	var p = new(UnionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = AzmParserRULE_union
	return p
}

func InitEmptyUnionContext(p *UnionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = AzmParserRULE_union
}

func (*UnionContext) IsUnionContext() {}

func NewUnionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *UnionContext {
	var p = new(UnionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = AzmParserRULE_union

	return p
}

func (s *UnionContext) GetParser() antlr.Parser { return s.parser }

func (s *UnionContext) AllPerm() []IPermContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IPermContext); ok {
			len++
		}
	}

	tst := make([]IPermContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IPermContext); ok {
			tst[i] = t.(IPermContext)
			i++
		}
	}

	return tst
}

func (s *UnionContext) Perm(i int) IPermContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPermContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPermContext)
}

func (s *UnionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *UnionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *UnionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case AzmVisitor:
		return t.VisitUnion(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *AzmParser) Union() (localctx IUnionContext) {
	localctx = NewUnionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, AzmParserRULE_union)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(35)
		p.Perm()
	}
	p.SetState(40)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == AzmParserT__0 {
		{
			p.SetState(36)
			p.Match(AzmParserT__0)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(37)
			p.Perm()
		}

		p.SetState(42)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IIntersectionContext is an interface to support dynamic dispatch.
type IIntersectionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllPerm() []IPermContext
	Perm(i int) IPermContext

	// IsIntersectionContext differentiates from other interfaces.
	IsIntersectionContext()
}

type IntersectionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyIntersectionContext() *IntersectionContext {
	var p = new(IntersectionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = AzmParserRULE_intersection
	return p
}

func InitEmptyIntersectionContext(p *IntersectionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = AzmParserRULE_intersection
}

func (*IntersectionContext) IsIntersectionContext() {}

func NewIntersectionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *IntersectionContext {
	var p = new(IntersectionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = AzmParserRULE_intersection

	return p
}

func (s *IntersectionContext) GetParser() antlr.Parser { return s.parser }

func (s *IntersectionContext) AllPerm() []IPermContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IPermContext); ok {
			len++
		}
	}

	tst := make([]IPermContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IPermContext); ok {
			tst[i] = t.(IPermContext)
			i++
		}
	}

	return tst
}

func (s *IntersectionContext) Perm(i int) IPermContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPermContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPermContext)
}

func (s *IntersectionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *IntersectionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *IntersectionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case AzmVisitor:
		return t.VisitIntersection(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *AzmParser) Intersection() (localctx IIntersectionContext) {
	localctx = NewIntersectionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, AzmParserRULE_intersection)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(43)
		p.Perm()
	}
	{
		p.SetState(44)
		p.Match(AzmParserT__1)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(45)
		p.Perm()
	}
	p.SetState(50)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == AzmParserT__1 {
		{
			p.SetState(46)
			p.Match(AzmParserT__1)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(47)
			p.Perm()
		}

		p.SetState(52)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IExclusionContext is an interface to support dynamic dispatch.
type IExclusionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllPerm() []IPermContext
	Perm(i int) IPermContext

	// IsExclusionContext differentiates from other interfaces.
	IsExclusionContext()
}

type ExclusionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExclusionContext() *ExclusionContext {
	var p = new(ExclusionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = AzmParserRULE_exclusion
	return p
}

func InitEmptyExclusionContext(p *ExclusionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = AzmParserRULE_exclusion
}

func (*ExclusionContext) IsExclusionContext() {}

func NewExclusionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExclusionContext {
	var p = new(ExclusionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = AzmParserRULE_exclusion

	return p
}

func (s *ExclusionContext) GetParser() antlr.Parser { return s.parser }

func (s *ExclusionContext) AllPerm() []IPermContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IPermContext); ok {
			len++
		}
	}

	tst := make([]IPermContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IPermContext); ok {
			tst[i] = t.(IPermContext)
			i++
		}
	}

	return tst
}

func (s *ExclusionContext) Perm(i int) IPermContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPermContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPermContext)
}

func (s *ExclusionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExclusionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ExclusionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case AzmVisitor:
		return t.VisitExclusion(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *AzmParser) Exclusion() (localctx IExclusionContext) {
	localctx = NewExclusionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, AzmParserRULE_exclusion)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(53)
		p.Perm()
	}
	{
		p.SetState(54)
		p.Match(AzmParserT__2)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(55)
		p.Perm()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IRelContext is an interface to support dynamic dispatch.
type IRelContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsRelContext differentiates from other interfaces.
	IsRelContext()
}

type RelContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyRelContext() *RelContext {
	var p = new(RelContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = AzmParserRULE_rel
	return p
}

func InitEmptyRelContext(p *RelContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = AzmParserRULE_rel
}

func (*RelContext) IsRelContext() {}

func NewRelContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *RelContext {
	var p = new(RelContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = AzmParserRULE_rel

	return p
}

func (s *RelContext) GetParser() antlr.Parser { return s.parser }

func (s *RelContext) CopyAll(ctx *RelContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *RelContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RelContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type DirectRelContext struct {
	RelContext
}

func NewDirectRelContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *DirectRelContext {
	var p = new(DirectRelContext)

	InitEmptyRelContext(&p.RelContext)
	p.parser = parser
	p.CopyAll(ctx.(*RelContext))

	return p
}

func (s *DirectRelContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DirectRelContext) ID() antlr.TerminalNode {
	return s.GetToken(AzmParserID, 0)
}

func (s *DirectRelContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case AzmVisitor:
		return t.VisitDirectRel(s)

	default:
		return t.VisitChildren(s)
	}
}

type SubjectRelContext struct {
	RelContext
}

func NewSubjectRelContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SubjectRelContext {
	var p = new(SubjectRelContext)

	InitEmptyRelContext(&p.RelContext)
	p.parser = parser
	p.CopyAll(ctx.(*RelContext))

	return p
}

func (s *SubjectRelContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SubjectRelContext) AllID() []antlr.TerminalNode {
	return s.GetTokens(AzmParserID)
}

func (s *SubjectRelContext) ID(i int) antlr.TerminalNode {
	return s.GetToken(AzmParserID, i)
}

func (s *SubjectRelContext) HASH() antlr.TerminalNode {
	return s.GetToken(AzmParserHASH, 0)
}

func (s *SubjectRelContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case AzmVisitor:
		return t.VisitSubjectRel(s)

	default:
		return t.VisitChildren(s)
	}
}

type WildcardRelContext struct {
	RelContext
}

func NewWildcardRelContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *WildcardRelContext {
	var p = new(WildcardRelContext)

	InitEmptyRelContext(&p.RelContext)
	p.parser = parser
	p.CopyAll(ctx.(*RelContext))

	return p
}

func (s *WildcardRelContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *WildcardRelContext) ID() antlr.TerminalNode {
	return s.GetToken(AzmParserID, 0)
}

func (s *WildcardRelContext) COLON() antlr.TerminalNode {
	return s.GetToken(AzmParserCOLON, 0)
}

func (s *WildcardRelContext) ASTERISK() antlr.TerminalNode {
	return s.GetToken(AzmParserASTERISK, 0)
}

func (s *WildcardRelContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case AzmVisitor:
		return t.VisitWildcardRel(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *AzmParser) Rel() (localctx IRelContext) {
	localctx = NewRelContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, AzmParserRULE_rel)
	p.SetState(64)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 4, p.GetParserRuleContext()) {
	case 1:
		localctx = NewDirectRelContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(57)
			p.Match(AzmParserID)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		localctx = NewWildcardRelContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(58)
			p.Match(AzmParserID)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(59)
			p.Match(AzmParserCOLON)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(60)
			p.Match(AzmParserASTERISK)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		localctx = NewSubjectRelContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(61)
			p.Match(AzmParserID)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(62)
			p.Match(AzmParserHASH)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(63)
			p.Match(AzmParserID)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IPermContext is an interface to support dynamic dispatch.
type IPermContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsPermContext differentiates from other interfaces.
	IsPermContext()
}

type PermContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPermContext() *PermContext {
	var p = new(PermContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = AzmParserRULE_perm
	return p
}

func InitEmptyPermContext(p *PermContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = AzmParserRULE_perm
}

func (*PermContext) IsPermContext() {}

func NewPermContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PermContext {
	var p = new(PermContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = AzmParserRULE_perm

	return p
}

func (s *PermContext) GetParser() antlr.Parser { return s.parser }

func (s *PermContext) CopyAll(ctx *PermContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *PermContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PermContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type ArrowPermContext struct {
	PermContext
}

func NewArrowPermContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ArrowPermContext {
	var p = new(ArrowPermContext)

	InitEmptyPermContext(&p.PermContext)
	p.parser = parser
	p.CopyAll(ctx.(*PermContext))

	return p
}

func (s *ArrowPermContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ArrowPermContext) AllID() []antlr.TerminalNode {
	return s.GetTokens(AzmParserID)
}

func (s *ArrowPermContext) ID(i int) antlr.TerminalNode {
	return s.GetToken(AzmParserID, i)
}

func (s *ArrowPermContext) ARROW() antlr.TerminalNode {
	return s.GetToken(AzmParserARROW, 0)
}

func (s *ArrowPermContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case AzmVisitor:
		return t.VisitArrowPerm(s)

	default:
		return t.VisitChildren(s)
	}
}

type DirectPermContext struct {
	PermContext
}

func NewDirectPermContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *DirectPermContext {
	var p = new(DirectPermContext)

	InitEmptyPermContext(&p.PermContext)
	p.parser = parser
	p.CopyAll(ctx.(*PermContext))

	return p
}

func (s *DirectPermContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DirectPermContext) ID() antlr.TerminalNode {
	return s.GetToken(AzmParserID, 0)
}

func (s *DirectPermContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case AzmVisitor:
		return t.VisitDirectPerm(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *AzmParser) Perm() (localctx IPermContext) {
	localctx = NewPermContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, AzmParserRULE_perm)
	p.SetState(70)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 5, p.GetParserRuleContext()) {
	case 1:
		localctx = NewDirectPermContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(66)
			p.Match(AzmParserID)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		localctx = NewArrowPermContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(67)
			p.Match(AzmParserID)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(68)
			p.Match(AzmParserARROW)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(69)
			p.Match(AzmParserID)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}
