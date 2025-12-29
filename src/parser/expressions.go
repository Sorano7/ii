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
		p.panic("No prefix found for: %s", p.currentToken().Type.String())
	}
	left := prefix()
	if left == nil {
		p.panic("Missing expression")
	}

	for !p.nextTokenIs(lexer.EOF) && prec < p.currentPrec() {
		infix := p.infixParseFns[p.currentToken().Type]
		if infix == nil {
			return left
		}
		left = infix(left)
		if left == nil {
			p.panic("Missing lhs expression")
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
		p.panic("Invalid number literal")
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

func (p *parser) parseStringLiteral() ast.Expression {
	s := &ast.StringLiteral{Value: p.currentToken().Value}
	p.advance()
	return s
}

func (p *parser) parsePrefixExpression() ast.Expression {
	expr := &ast.PrefixExpression{
		Operator: p.currentToken().Value,
	}

	p.advance()
	expr.Right = p.parseExpression(Prefix)
	if expr.Right == nil {
		p.panic("Missing rhs expression for prefix")
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
		p.panic("Missing rhs expression for infix")
	}

	return expr
}

func (p *parser) parseGroupedExpression() ast.Expression {
	p.advanceOrPanic(lexer.LParen)
	if p.currentTokenIs(lexer.RParen) {
		p.panic("Empty grouped expression")
	}

	exp := p.parseExpression(Lowest)
	if exp == nil {
		p.panic("Missing expression in ()")
	}

	p.advanceOrPanic(lexer.RParen)
	return exp
}

func (p *parser) parsePair() ast.Expression {
	if p.currentTokenIs(lexer.RBracket) {
		p.advance()
		return makeNil()
	}
	
	first := p.parseExpression(Lowest)
	if first == nil {
		p.panic("Missing first expression in pair")
	}

	if p.currentTokenIs(lexer.RBracket) {
		p.advance()
		return &ast.PairLiteral{
			First: first,
			Second: makeNil(),
		}
	} 

	if p.currentTokenIs(lexer.Bar) {
		return p.parseFCompExpression(first)
	}

	p.advanceOrPanic(lexer.Comma)

	second := p.parseExpression(Lowest)
	if second == nil {
		p.panic("Missing second expression in pair")
	}

	p.advanceOrPanic(lexer.RBracket)
	
	return &ast.PairLiteral{
		First: first,
		Second: second,
	}
}

func (p *parser) parseFunctionOrPair() ast.Expression {
	p.advanceOrPanic(lexer.LBracket)

	if p.currentTokenIs(lexer.ComparisonOps...) || p.currentTokenIs(lexer.ArithmeticOps...) {
		return p.parseShortInfixFunction()
	}

	if !p.nextTokenIs(lexer.Lambda) {
		return p.parsePair()
	}

	param := p.currentToken().Value
	p.advance()
	p.advance()

	body := p.parseExpression(Lowest)
	if body == nil {
		p.panic("Missing anonymous function body")
	}

	p.advanceOrPanic(lexer.RBracket)

	return &ast.FunctionLiteral{
		Parameter: param,
		Body:      body,
	}
}

func (p *parser) parseShortInfixFunction() ast.Expression {
	op := p.currentToken().Value
	p.advance()

	body := p.parseExpression(Lowest)
	if body == nil {
		p.panic("Missing short infix function body")
	}

	p.advanceOrPanic(lexer.RBracket)

	return &ast.FunctionLiteral{
		Parameter: "x",
		Body: &ast.InfixExpression{
			Left: &ast.Identifier{Value: "x"},
			Operator: op,
			Right: body,
		},
	}
}

func (p *parser) parseFCompExpression(first ast.Expression) ast.Expression {
	p.advanceOrPanic(lexer.Bar)

	second := p.parseExpression(Lowest)
	if second == nil {
		p.panic("Missing second expression for fcomp")
	}

	p.advanceOrPanic(lexer.RBracket)

	return &ast.FCompExpression{
		First: first,
		Second: second,
	}
}

func (p *parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Function: function}
	p.advanceOrPanic(lexer.LParen)

	if p.currentTokenIs(lexer.RParen) {
		p.advance()
		return exp
	}

	exp.Argument = p.parseExpression(Lowest)
	if exp.Argument == nil {
		p.panic("Missing call expression argument")
	}

	p.advanceOrPanic(lexer.RParen)

	return exp
}

func (p *parser) parseLambdaExpression(left ast.Expression) ast.Expression {
	ident, ok := left.(*ast.Identifier)
	if !ok {
		p.panic("Expects identifier, found %T", left)
	}

	param := ident.Value

	p.advance()
	body := p.parseExpression(Lowest)
	if body == nil {
		p.panic("Missing lambda function body")
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
	case "<<", ".":
		prec++
	case "<>":
		prec--
	}
	expr.Right = p.parseExpression(prec)

	if expr.Right == nil {
		p.panic("Missing rhs expression for pipe operator")
	}

	return expr
}


func (p *parser) parseCondExpression(left ast.Expression) ast.Expression {
	expr := &ast.CondExpression{If: left}
	prec := p.currentPrec()

	p.advanceOrPanic(lexer.If)

	expr.Then = p.parseExpression(prec - 1)
	if expr.Then == nil {
		p.panic("Missing then clause expression")
	}

	if !p.currentTokenIs(lexer.Else) {
		expr.Else = makeNil()
		return expr
	}
	p.advance()

	expr.Else = p.parseExpression(prec - 1)
	if expr.Else == nil {
		p.panic("Missing else clause expression")
	}
	return expr
}

func (p *parser) parseArrayLiteral() ast.Expression {
	p.advanceOrPanic(lexer.LCurly)

	if p.currentTokenIs(lexer.RCurly) {
		p.advance()
		return makeNil()
	}

	array := &ast.ArrayLiteral{Elements: make([]ast.Expression, 0)}

	head := p.parseExpression(Lowest)
	if head == nil {
		p.panic("Missing first element in array")
	}

	if p.currentTokenIs(lexer.Ellipsis) {
		return p.parseSequence(head)
	}

	array.Elements = append(array.Elements, head)
	
	if p.currentTokenIs(lexer.Comma) {
		p.advance()
	}

	for !p.currentTokenIs(lexer.RCurly) {
		expr := p.parseExpression(Lowest)
		if expr == nil {
			p.panic("Missing expression in array")
		}
		array.Elements = append(array.Elements, expr)

		if p.currentTokenIs(lexer.Comma) {
            p.advance()
            continue
        }

		if p.currentTokenIs(lexer.RCurly) {
            break
        }
		p.panic("Missing comma in array")
	}

	p.advanceOrPanic(lexer.RCurly)

	return array
}

func (p *parser) parseSequence(head ast.Expression) ast.Expression {
	a, ok := head.(*ast.NumberLiteral)
	if !ok {
		p.panic("Expects number, found: %T", head)
	}

	p.advanceOrPanic(lexer.Ellipsis)

	b, ok := p.parseExpression(Lowest).(*ast.NumberLiteral)
	if !ok {
		p.panic("Expects number, found: %T", head)
	}

	p.advanceOrPanic(lexer.RCurly)

	return &ast.CallExpression{
		Function: &ast.CallExpression{
			Function: &ast.Identifier{Value: "seq"},
			Argument: a,
		},
		Argument: b,
	}
}