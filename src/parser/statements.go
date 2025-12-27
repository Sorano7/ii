package parser

import (
	"lambda/src/ast"
	"lambda/src/lexer"
)

func (p *parser) parseStatement() ast.Statement {
	if !p.currentTokenIs(lexer.Identifier) || !p.nextTokenIs(lexer.Colon) {
		return p.parseExpressionStatement()
	}
	return p.parseBindingStatement()
}

func (p *parser) parseBindingStatement() *ast.BindingStatement {
	stmt := &ast.BindingStatement{
		Name: p.currentToken().Value,
	}
	if !p.nextTokenIs(lexer.Colon) {
		return nil
	}

	p.advance()
	p.advance()

	stmt.Value = p.parseExpression(Lowest)
	if stmt.Value == nil {
		return nil
	}

	return stmt
}

func (p *parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{}
	stmt.Expression = p.parseExpression(Lowest)
	if stmt.Expression == nil {
		return nil
	}

	return stmt
}