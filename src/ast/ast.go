package ast

import (
	"strings"
)

type Node interface {
	String() string
	Debug() string
}

type Expression interface {
	expressionNode()
	Node
}

type Statement interface {
	statementNode()
	Node
}

type Program struct {
	Statements []Statement
	Node
}

func (p *Program) String() string {
	var sb strings.Builder
	for _, s := range p.Statements {
		sb.WriteString(s.String())
		sb.WriteString("\n")
	}
	return sb.String()
}
func (p *Program) Debug() string {
	var sb strings.Builder
	for _, s := range p.Statements {
		sb.WriteString(s.Debug())
		sb.WriteString("\n")
	}
	return sb.String()
}