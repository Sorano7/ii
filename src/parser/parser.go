package parser

import (
	"fmt"
	"lambda/src/ast"
	"lambda/src/lexer"
)

type precedence int

const (
	_ precedence = iota
	Lowest
	Control
	Equals
	LessMore
	Sum
	Product
	Prefix
	Call
	Lambda
)

var precedences = map[lexer.TokenType]precedence{
	lexer.LPipe:         Control,
	lexer.RPipe:         Control,
	lexer.Diamond:       Control,
	lexer.If:            Control,
	lexer.Equals:        Equals,
	lexer.NotEquals:     Equals,
	lexer.Less:          LessMore,
	lexer.Greater:       LessMore,
	lexer.LessEquals:    LessMore,
	lexer.GreaterEquals: LessMore,
	lexer.Plus:          Sum,
	lexer.Dash:          Sum,
	lexer.Dot:           Sum,
	lexer.Star:          Product,
	lexer.Slash:         Product,
	lexer.Percent:       Product,
	lexer.LParen:        Call,
	lexer.Lambda:        Lambda,
}

type parser struct {
	tokens         []lexer.Token
	pos            int
	prefixParseFns map[lexer.TokenType]prefixParseFn
	infixParseFns  map[lexer.TokenType]infixParseFn
}

func (p *parser) currentToken() lexer.Token {
	return p.tokens[p.pos]
}

func (p *parser) advance() lexer.Token {
	tk := p.currentToken()
	p.pos++
	return tk
}

func (p *parser) peek(n int) lexer.Token {
	if p.currentToken().In(lexer.EOF) {
		return p.currentToken()
	}
	return p.tokens[p.pos+n]
}

func (p *parser) currentPrec() precedence {
	if p, ok := precedences[p.currentToken().Type]; ok {
		return p
	}
	return Lowest
}

func (p *parser) advanceOrPanic(t lexer.TokenType) {
	if p.currentTokenIs(t) {
		p.advance()
		return
	}
	p.panic("Expects %s, found (%s)", t.String(), p.currentToken().String())
}

func (p *parser) panic(format string, a...any) {
	msg := fmt.Sprintf(format, a...)
	panic(msg)
}

func (p *parser) currentTokenIs(t ...lexer.TokenType) bool {
	return p.currentToken().In(t...)
}

func (p *parser) nextTokenIs(t ...lexer.TokenType) bool {
	return p.peek(1).In(t...)
}

func (p *parser) registerPrefix(f prefixParseFn, ts ...lexer.TokenType) {
	for _, t := range ts {
		p.prefixParseFns[t] = f
	}
}

func (p *parser) registerInfix(f infixParseFn, ts ...lexer.TokenType) {
	for _, t := range ts {
		p.infixParseFns[t] = f
	}
}

func createParser(tokens []lexer.Token) *parser {
	p := &parser{
		tokens:         tokens,
		pos:            0,
		prefixParseFns: make(map[lexer.TokenType]prefixParseFn),
		infixParseFns:  make(map[lexer.TokenType]infixParseFn),
	}

	p.registerPrefix(p.parseIdentifier, lexer.Identifier)
	p.registerPrefix(p.parseNumberLiteral, lexer.Number)
	p.registerPrefix(p.parseCharLiteral, lexer.Char)
	p.registerPrefix(p.parseStringLiteral, lexer.String)
	p.registerPrefix(p.parsePrefixExpression, lexer.Dash)
	p.registerPrefix(p.parseGroupedExpression, lexer.LParen)
	p.registerPrefix(p.parseFunctionOrPair, lexer.LBracket)
	p.registerPrefix(p.parseArrayLiteral, lexer.LCurly)

	p.registerInfix(p.parseInfixExpression,
		lexer.Plus, lexer.Dash, lexer.Star, lexer.Slash, lexer.Percent,
		lexer.Equals, lexer.NotEquals, lexer.Less, lexer.LessEquals,
		lexer.Greater, lexer.GreaterEquals,
	)
	p.registerInfix(p.parseCallExpression, lexer.LParen)
	p.registerInfix(p.parsePipeExpression, lexer.LPipe, lexer.RPipe, lexer.Diamond, lexer.Dot)
	p.registerInfix(p.parseLambdaExpression, lexer.Lambda)
	p.registerInfix(p.parseCondExpression, lexer.If)

	return p
}

func Parse(src string) (*ast.Program, bool) {
	tokens := lexer.Tokenize(src)
	if tokens == nil {
		return nil, false
	}

	program := &ast.Program{Statements: []ast.Statement{}}
	p := createParser(tokens)

	for !p.currentTokenIs(lexer.EOF) {
		start := p.pos
		stmt := p.parseStatement()
		if stmt == nil {
			return nil, false
		}
		program.Statements = append(program.Statements, stmt)
		if p.pos == start {
			p.advance()
		}
	}

	if len(program.Statements) <= 0 {
		return nil, false
	}
	return program, true
}
