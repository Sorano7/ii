package parser

import (
	"lambda/src/ast"
	"lambda/src/lexer"
	"strconv"
)

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

func makeNil() ast.Expression {
	return &ast.Identifier{Value: "nil"}
}

func (p *parser) parseExpression(prec precedence) ast.Expression {
	prefix := p.prefixParseFns[p.currentToken().Type]
	if prefix == nil {
		return nil
	}
	left := prefix()
	if left == nil {
		return nil
	}

	for !p.nextTokenIs(lexer.EOF) && prec < p.currentPrec() {
		infix := p.infixParseFns[p.currentToken().Type]
		if infix == nil {
			return left
		}
		left = infix(left)
		if left == nil {
			return nil
		}
	}

	return left
}

func (p *parser) parseIdentifier() ast.Expression {
	ident := &ast.Identifier{Value: p.currentToken().Value}
	p.advance()
	return ident
}

func (p *parser) parseNumberLiteral() ast.Expression {
	value, err := strconv.ParseFloat(p.currentToken().Value, 64)
	if err != nil {
		return nil
	}

	num := &ast.NumberLiteral{Value: value}
	p.advance()
	return num
}

func (p *parser) parseCharLiteral() ast.Expression {
	c := &ast.CharLiteral{Value: p.currentToken().Value}
	p.advance()
	return c
}

func (p *parser) parsePrefixExpression() ast.Expression {
	expr := &ast.PrefixExpression{
		Operator: p.currentToken().Value,
	}

	p.advance()
	expr.Right = p.parseExpression(Prefix)
	if expr.Right == nil {
		return nil
	}

	return expr
}

func (p *parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expr := &ast.InfixExpression{
		Operator: p.currentToken().Value,
		Left:     left,
	}

	prec := p.currentPrec()
	p.advance()
	expr.Right = p.parseExpression(prec)
	if expr.Right == nil {
		return nil
	}

	return expr
}

func (p *parser) parseGroupedExpression() ast.Expression {
	if _, ok := p.consume(lexer.LParen); !ok {
		return nil
	}

	if p.currentTokenIs(lexer.RParen) {
		return nil
	}

	exp := p.parseExpression(Lowest)
	if exp == nil {
		return nil
	}

	if _, ok := p.consume(lexer.RParen); !ok {
		return nil
	}
	return exp
}

func (p *parser) parsePair() ast.Expression {
	if p.currentTokenIs(lexer.RBracket) {
		p.advance()
		return makeNil()
	}
	
	first := p.parseExpression(Lowest)
	if first == nil {
		return nil
	}

	if p.currentTokenIs(lexer.RBracket) {
		p.advance()
		return &ast.PairLiteral{
			First: first,
			Second: makeNil(),
		}
	} 

	p.advance()

	second := p.parseExpression(Lowest)
	if second == nil {
		return nil
	}

	if _, ok := p.consume(lexer.RBracket); !ok {
		return nil
	}
	
	return &ast.PairLiteral{
		First: first,
		Second: second,
	}
}

func (p *parser) parseLambdaOrPair() ast.Expression {
	if _, ok := p.consume(lexer.LBracket); !ok {
		return nil
	}

	if !p.currentTokenIs(lexer.Identifier) || !p.nextTokenIs(lexer.Lambda) {
		return p.parsePair()
	}

	param := p.currentToken().Value
	p.advance()
	p.advance()

	body := p.parseExpression(Lowest)
	if body == nil {
		return nil
	}

	if _, ok := p.consume(lexer.RBracket); !ok {
		return nil
	}

	return &ast.FunctionLiteral{
		Parameter: param,
		Body:      body,
	}
}

func (p *parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Function: function}
	if _, ok := p.consume(lexer.LParen); !ok {
		return nil
	}

	if p.currentTokenIs(lexer.RParen) {
		p.advance()
		return exp
	}

	exp.Argument = p.parseExpression(Lowest)
	if exp.Argument == nil {
		return nil
	}

	if _, ok := p.consume(lexer.RParen); !ok {
		return nil
	}

	return exp
}

func (p *parser) parseLambdaExpression(left ast.Expression) ast.Expression {
	ident, ok := left.(*ast.Identifier)
	if !ok {
		return nil
	}

	param := ident.Value

	p.advance()
	body := p.parseExpression(Lowest)
	if body == nil {
		return nil
	}

	return &ast.FunctionLiteral{
		Parameter: param,
		Body:      body,
	}
}

func (p *parser) parsePipeExpression(left ast.Expression) ast.Expression {
	expr := &ast.PipeExpression{
		Direction: p.currentToken().Value,
		Left:      left,
	}

	prec := p.currentPrec()
	p.advance()

	switch expr.Direction {
	case "<<":
		prec++
	case "<>":
		prec--
	}
	expr.Right = p.parseExpression(prec)

	if expr.Right == nil {
		return nil
	}

	return expr
}

func (p *parser) parseArrayLiteral() ast.Expression {
	if _, ok := p.consume(lexer.LCurly); !ok {
		return nil
	}
	if p.currentTokenIs(lexer.RCurly) {
		p.advance()
		return makeNil()
	}

	array := &ast.ArrayLiteral{Elements: make([]ast.Expression, 0)}

	for !p.currentTokenIs(lexer.RCurly) {
		expr := p.parseExpression(Lowest)
		if expr == nil {
			return nil
		}
		array.Elements = append(array.Elements, expr)

		if p.currentTokenIs(lexer.Comma) {
            p.advance()
            continue
        }

		if p.currentTokenIs(lexer.RCurly) {
            break
        }
		return nil
	}

	if _, ok := p.consume(lexer.RCurly); !ok {
        return nil
    }

	return array
}