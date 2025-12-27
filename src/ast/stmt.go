package ast

import "fmt"

type BindingStatement struct {
	Name  string
	Value Expression
}

func (b *BindingStatement) statementNode() {}
func (b *BindingStatement) String() string {
	return fmt.Sprintf("%s: %s", b.Name, b.Value.String())
}
func (b *BindingStatement) Debug() string {
	return fmt.Sprintf("{ Type: Binding, Identifer: %s, Expression: %s }",
		b.Name, b.Value.Debug())
}

type ExpressionStatement struct {
	Expression Expression
}

func (e *ExpressionStatement) statementNode() {}
func (e *ExpressionStatement) String() string {
	return e.Expression.String()
}
func (e *ExpressionStatement) Debug() string {
	return e.Expression.Debug()
}
