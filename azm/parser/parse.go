package parser

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/antlr4-go/antlr/v4"
	"github.com/aserto-dev/topaz/azm/model"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

var (
	ErrIdentifierExpected = errors.New("identifier expected")
	ErrInvalidExpression  = errors.New("invalid expression")
	ErrParse              = errors.New("parse error")
)

func ParseRelation(input string) ([]*model.RelationRef, error) {
	p := newParser(input)

	rTree, err := p.Relation(), p.Error()
	if err != nil {
		return nil, err
	}

	var v RelationVisitor

	return v.Visit(rTree).([]*model.RelationRef), nil
}

func ParsePermission(input string) (*model.Permission, error) {
	p := newParser(input)

	pTree, err := p.Permission(), p.Error()
	if err != nil {
		return nil, err
	}

	var v PermissionVisitor

	return v.Visit(pTree).(*model.Permission), nil
}

func newParser(input string) *parser {
	lexer := NewAzmLexer(antlr.NewInputStream(input))
	stream := antlr.NewCommonTokenStream(lexer, 0)
	listener := newErrorListener(input)

	p := NewAzmParser(stream)
	p.AddErrorListener(listener)

	if os.Getenv("AZM_DIAGNOSTICS") == "1" {
		p.AddErrorListener(antlr.NewDiagnosticErrorListener(true))
	}

	return &parser{AzmParser: p, listener: listener}
}

type parser struct {
	*AzmParser

	listener *errorListener
}

func (p *parser) Error() error {
	return p.listener.err
}

type errorListener struct {
	*antlr.DefaultErrorListener

	input string
	err   error
}

func newErrorListener(input string) *errorListener {
	return &errorListener{antlr.NewDefaultErrorListener(), input, nil}
}

func (l *errorListener) SyntaxError(
	_ antlr.Recognizer, _ any,
	line, column int,
	msg string,
	e antlr.RecognitionException,
) {
	if e != nil {
		// lexer recognition error
		matches := recognitionErrorRegexp().FindStringSubmatch(msg)
		if len(matches) == recognitionRegexpMatches {
			actual, expected := matches[1], matches[2]
			message := fmt.Sprintf("unexpected %s in '%s'. expected %s", actual, l.input, expected)
			l.err = multierror.Append(l.err, errors.Wrap(ErrInvalidExpression, message))

			return
		}
	}

	if e == nil && strings.HasPrefix(msg, "extraneous input") {
		// extraneous input parser error
		symbol := l.input[column : column+1]
		message := fmt.Sprintf("unexpected '%s' in '%s'", symbol, l.input)
		l.err = multierror.Append(l.err, errors.Wrap(ErrIdentifierExpected, message))

		return
	}

	l.err = multierror.Append(l.err, errors.Wrap(ErrParse, msg))
}

var recognitionErrorRegexp = sync.OnceValue(func() *regexp.Regexp {
	return regexp.MustCompile(`^mismatched input ('.+') expecting (.+)$`)
})

const recognitionRegexpMatches = 3
