package ast

import (
	"fmt"
	"strings"
)

type Identifier struct {
	Value string
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) String() string  { return i.Value }
func (i *Identifier) Debug() string {
	return fmt.Sprintf("{ Type: Identifier, Value: %s }", i.Value)
}

type NumberLiteral struct {
	Value float64
}

func (n *NumberLiteral) expressionNode() {}
func (n *NumberLiteral) String() string  { return fmt.Sprint(n.Value) }
func (n *NumberLiteral) Debug() string {
	return fmt.Sprintf("{ Type: Number, Value: %f }", n.Value)
}

type CharLiteral struct {
	Value string
}

func (c *CharLiteral) expressionNode() {}
func (c *CharLiteral) String() string  { return c.Value }
func (c *CharLiteral) Debug() string {
	return fmt.Sprintf("{ Type: Char, Value: %s }", c.Value)
}

type StringLiteral struct {
	Value string
}

func (s *StringLiteral) expressionNode() {}
func (s *StringLiteral) String() string  { return s.Value }
func (s *StringLiteral) Debug() string {
	return fmt.Sprintf("{ Type: String, Value: %s }", s.Value)
}

type PairLiteral struct {
	First  Expression
	Second Expression
}

func (p *PairLiteral) expressionNode() {}
func (p *PairLiteral) String() string {
	return fmt.Sprintf("[%s, %s]", p.First.String(), p.Second.String())
}
func (p *PairLiteral) Debug() string {
	return fmt.Sprintf("{ Type: Pair, First: %s, Second: %s }",
		p.First.Debug(), p.Second.Debug())
}

type ArrayLiteral struct {
	Elements []Expression	
}

func (a *ArrayLiteral) expressionNode()	{}
func (a *ArrayLiteral) String() string {
	elementStrings := make([]string, 0)
	for _, e := range a.Elements {
		elementStrings = append(elementStrings, e.String())
	}
	return fmt.Sprintf("{ %s }", strings.Join(elementStrings, ", "))
}
func (a *ArrayLiteral) Debug() string {
	elementStrings := make([]string, 0)
	for _, e := range a.Elements {
		elementStrings = append(elementStrings, e.Debug())
	}
	return fmt.Sprintf("{ Type: Array, Elements: { %s } }", strings.Join(elementStrings, ", "))
}

type FunctionLiteral struct {
	Parameter string
	Body      Expression
}

func (f *FunctionLiteral) expressionNode() {}
func (f *FunctionLiteral) String() string {
	return fmt.Sprintf("[%s; %s]", f.Parameter, f.Body.String())
}
func (f *FunctionLiteral) Debug() string {
	return fmt.Sprintf("{ Type: Function, Parameter: %s, Body: %s }",
		f.Parameter, f.Body.Debug())
}

type CallExpression struct {
	Function Expression
	Argument Expression
}

func (c *CallExpression) expressionNode() {}
func (c *CallExpression) String() string {
	return fmt.Sprintf("%s (%s)", c.Function.String(), c.Argument.String())
}
func (c *CallExpression) Debug() string {
	arg := ""
	if c.Argument != nil {
		arg = c.Argument.String()
	}
	return fmt.Sprintf("{ Type: Call, Function: %s, Argument: %s }",
		c.Function.Debug(), arg)
}

type InfixExpression struct {
	Left     Expression
	Operator string
	Right    Expression
}

func (i *InfixExpression) expressionNode() {}
func (i *InfixExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", i.Left.String(), i.Operator, i.Right.String())
}
func (i *InfixExpression) Debug() string {
	return fmt.Sprintf("{ Type: Infix, Left: %s, Operator: %s, Right: %s }",
		i.Left.Debug(), i.Operator, i.Right.Debug())
}

type PrefixExpression struct {
	Operator string
	Right    Expression
}

func (p *PrefixExpression) expressionNode() {}
func (p *PrefixExpression) String() string {
	return fmt.Sprintf("%s%s", p.Operator, p.Right.String())
}
func (p *PrefixExpression) Debug() string {
	return fmt.Sprintf("{ Type: Prefix, Operator: %s, Right: %s }",
		p.Operator, p.Right.Debug())
}

type PipeExpression struct {
	Left      Expression
	Right     Expression
	Direction string
}

func (p *PipeExpression) expressionNode() {}
func (p *PipeExpression) String() string {
	return fmt.Sprintf("%s %s %s", p.Left.String(), p.Direction, p.Right.String())
}
func (p *PipeExpression) Debug() string {
	return fmt.Sprintf("{ Type: Pipe, Left: %s, Direction: %s, Right: %s }",
		p.Left.Debug(), p.Direction, p.Right.Debug())
}

type CondExpression struct {
	If Expression
	Then Expression
	Else Expression
}

func (c *CondExpression) expressionNode() {}
func (c *CondExpression) String() string {
	return fmt.Sprintf("(%s) ?? (%s) !! (%s)", c.If.String(), c.Then.String(), c.Else.String())
}
func (c *CondExpression) Debug() string {
	return fmt.Sprintf("{ Type: Cond, If: %s, Else: %s, Then: %s}", 
		c.If.Debug(), c.Then.Debug(), c.Else.Debug())	
}