package parser

import (
	"lambda/src/ast"
	"lambda/src/lexer"
)

func (p *parser) parseStatement() ast.Statement {
	if !p.currentTokenIs(lexer.Identifier) || !p.nextTokenIs(lexer.Assignment) {
		return p.parseExpressionStatement()
	}
	return p.parseBindingStatement()
}

func (p *parser) parseBindingStatement() *ast.BindingStatement {
	stmt := &ast.BindingStatement{Name: p.currentToken().Value}
	p.advanceOrPanic(lexer.Identifier)
	p.advanceOrPanic(lexer.Assignment)

	if p.currentTokenIs(lexer.EOF) {
		p.panic("Missing assignment body")
	}

	stmt.Value = p.parseExpression(Lowest)
	if stmt.Value == nil {
		p.panic("Missing assignment body")
	}

	return stmt
}

func (p *parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{}
	stmt.Expression = p.parseExpression(Lowest)
	if stmt.Expression == nil {
		p.panic("Missing expression")
	}

	return stmt
}