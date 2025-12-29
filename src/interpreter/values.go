package interpreter

import (
	"fmt"
	"lambda/src/ast"
	"strings"
)

type ValueType string

const (
	BoolValue      ValueType = "Bool"
	BoolThunkValue ValueType = "BoolThunk"
	NumberValue    ValueType = "Number"
	CharValue      ValueType = "Char"
	PairValue      ValueType = "Pair"
	ArrayValue     ValueType = "Array"
	StringValue    ValueType = "String"
	FunctionValue  ValueType = "Function"
	NilValue       ValueType = "Nil"
	ErrorValue     ValueType = "Error"
	UndefinedValue ValueType = "Undefined"
	ThunkValue     ValueType = "Thunk"
	BuiltinValue   ValueType = "Built-in"
)

type Value interface {
	Type() ValueType
	String() string
}

type Bool struct {
	Value bool
}

func (b *Bool) Type() ValueType { return BoolValue }
func (b *Bool) String() string {
	if b.Value {
		return "true"
	}
	return "false"
}

type BoolThunk struct {
	Cond  bool
	First Value
}

func (b *BoolThunk) Type() ValueType { return BoolThunkValue }
func (b *BoolThunk) String() string {
	if b.Cond {
		return fmt.Sprintf("[x; %s]", b.First.String())
	}
	return "[x; x]"
}

type Number struct {
	Value float64
}

func (n *Number) Type() ValueType { return NumberValue }
func (n *Number) String() string {
	return fmt.Sprint(n.Value)
}

type Char struct {
	Value string
}

func (c *Char) Type() ValueType { return CharValue }
func (c *Char) String() string {
	return "'" + c.Value + "'"
}

type Pair struct {
	First  Value
	Second Value
	Env    *Environment
}

func (p *Pair) Type() ValueType { return PairValue }
func (p *Pair) String() string {
	e := createEvaluator()
	fst := e.force(p.First)
	if isError(fst) {
		return fst.String()
	}
	snd := e.force(p.Second)
	if isError(snd) {
		return snd.String()
	}
	return fmt.Sprintf("[%s, %s]", fst.String(), snd.String())
}

type Array struct {
	Elements []Value
}

func (a *Array) Type() ValueType { return ArrayValue }
func (a *Array) String() string {
	e := createEvaluator()
	elStrings := make([]string, 0)
	for _, el := range a.Elements {
		forced := e.force(el)
		if isError(forced) {
			return forced.String()
		}
		elStrings = append(elStrings, forced.String())
	}
	return fmt.Sprintf("{ %s }", strings.Join(elStrings, ", "))
}

type String struct {
	Value string
}

func (s *String) Type() ValueType { return StringValue }
func (s *String) String() string  { return "\"" + s.Value + "\"" }

type Nil struct{}

func (n *Nil) Type() ValueType { return NilValue }
func (n *Nil) String() string  { return "nil" }

type Function struct {
	Name      string
	Parameter string
	Body      ast.Expression
	Env       *Environment
}

func (f *Function) Type() ValueType { return FunctionValue }
func (f *Function) String() string {
	return f.renderString(make(map[*Function]bool))
}
func (f *Function) renderString(seen map[*Function]bool) string {
	if seen == nil {
		seen = make(map[*Function]bool)
	}
	seen[f] = true
	body := renderExpressionWithEnv(f.Body, f.Env, seen)
	delete(seen, f)
	return fmt.Sprintf("[%s; %s]", f.Parameter, body)
}

func renderExpressionWithEnv(expr ast.Expression, env *Environment, seen map[*Function]bool) string {
	if expr == nil {
		return ""
	}

	renderEval := createEvaluator()

	switch node := expr.(type) {
	case *ast.NumberLiteral:
		return fmt.Sprint(node.Value)

	case *ast.CharLiteral:
		return node.Value

	case *ast.Identifier:
		if env == nil {
			return node.String()
		}

		if val, ok := env.Get(node.Value); ok && val != nil {
			switch v := val.(type) {
			case *Number, *Char, *Error, *Undefined:
				return v.String()
			case *Function:
				if seen == nil {
					seen = make(map[*Function]bool)
				}
				if seen[v] {
					return node.String()
				}
				seen[v] = true
				s := v.renderString(seen)
				delete(seen, v)
				return s

			case *Thunk:
				if v.bindingName != "" && v.bindingName == node.Value {
					if !v.evaluated {
						return node.String()
					}
					if _, ok := v.value.(*Function); ok {
						return node.String()
					}
				}

				if !v.evaluated {
					forced := v.Force(renderEval)
					return forced.String()
				}

				if v.evaluated && v.value != nil {
					switch inner := v.value.(type) {
					case *Number, *Char, *Error, *Undefined:
						return inner.String()
					case *Function:
						if seen == nil {
							seen = make(map[*Function]bool)
						}
						if seen[inner] {
							return node.String()
						}
						seen[inner] = true
						s := inner.renderString(seen)
						delete(seen, inner)
						return s
					default:
						return inner.String()
					}
				}

				return v.String()

			default:
				return val.String()
			}
		}

	case *ast.InfixExpression:
		left := renderExpressionWithEnv(node.Left, env, seen)
		right := renderExpressionWithEnv(node.Right, env, seen)
		return fmt.Sprintf("(%s %s %s)", left, node.Operator, right)

	case *ast.PrefixExpression:
		right := renderExpressionWithEnv(node.Right, env, seen)
		return fmt.Sprintf("%s%s", node.Operator, right)

	case *ast.FunctionLiteral:
		body := renderExpressionWithEnv(node.Body, env, seen)
		return fmt.Sprintf("[%s; %s]", node.Parameter, body)

	case *ast.CallExpression:
		fn := renderExpressionWithEnv(node.Function, env, seen)
		if node.Argument == nil {
			return fmt.Sprintf("%s", fn)
		}
		arg := renderExpressionWithEnv(node.Argument, env, seen)
		return fmt.Sprintf("%s (%s)", fn, arg)

	case *ast.PairLiteral:
		first := renderExpressionWithEnv(node.First, env, seen)
		second := renderExpressionWithEnv(node.Second, env, seen)
		return fmt.Sprintf("[%s, %s]", first, second)

	case *ast.PipeExpression:
		left := renderExpressionWithEnv(node.Left, env, seen)
		right := renderExpressionWithEnv(node.Right, env, seen)
		return fmt.Sprintf("%s %s %s", left, node.Direction, right)
	}

	return expr.String()
}

type Error struct {
	Message string
}

func (e *Error) Type() ValueType { return ErrorValue }
func (e *Error) String() string {
	return fmt.Sprintf("[Error] %s", e.Message)
}

func error(format string, a ...any) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

func isError(v Value) bool {
	return v != nil && v.Type() == ErrorValue
}

type Undefined struct{}

func (u *Undefined) Type() ValueType { return UndefinedValue }
func (u *Undefined) String() string {
	return "undefined"
}

type Thunk struct {
	expr        ast.Expression
	env         *Environment
	evaluated   bool
	value       Value
	bindingName string
}

func (t *Thunk) Type() ValueType { return ThunkValue }
func (t *Thunk) String() string {
	if !t.evaluated {
		t.value = createEvaluator().Evaluate(t.expr, t.env)
	}
	if t.value == nil {
		return "undefined"
	}
	return t.value.String()
}

func (t *Thunk) Force(e *Evaluator) Value {
	if t.evaluated {
		return t.value
	}
	if t.expr == nil {
		t.value = &Undefined{}
	} else {
		t.value = e.Evaluate(t.expr, t.env)
		if t.bindingName != "" {
			if fn, ok := t.value.(*Function); ok {
				if fn.Name == "" {
					fn.Name = t.bindingName
				}
			}
		}
	}
	t.evaluated = true
	return t.value
}

type BuiltinFunction func(arg Value) Value
type Builtin struct {
	Name string
	Fn   BuiltinFunction
}

func (b *Builtin) Type() ValueType { return BuiltinValue }
func (b *Builtin) String() string  { return fmt.Sprintf("<builtin:%s>", b.Name) }