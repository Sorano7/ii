package interpreter

import (
	"lambda/src/ast"
	"math"
)

type Evaluator struct{}

func createEvaluator() *Evaluator {
	return &Evaluator{}
}

func (e *Evaluator) Evaluate(node ast.Node, env *Environment) Value {
	switch node := node.(type) {
	case *ast.Program:
		return e.evalProgram(node, env)

	case *ast.BindingStatement:
		return e.evalBinding(node, env)

	case *ast.ExpressionStatement:
		return e.Evaluate(node.Expression, env)

	case *ast.NumberLiteral:
		return &Number{Value: node.Value}

	case *ast.CharLiteral:
		return &Char{Value: node.Value}

	case *ast.Identifier:
		return e.evalIdentifier(node, env)

	case *ast.PrefixExpression:
		right := e.Evaluate(node.Right, env)
		if isError(right) {
			return right
		}
		return e.evalPrefix(node.Operator, right)

	case *ast.InfixExpression:
		left := e.Evaluate(node.Left, env)
		if isError(left) {
			return left
		}
		right := e.Evaluate(node.Right, env)
		if isError(right) {
			return right
		}
		return e.evalInfix(node.Operator, left, right, env)

	case *ast.FunctionLiteral:
		return &Function{
			Parameter: node.Parameter,
			Body:      node.Body,
			Env:       env,
		}

	case *ast.CallExpression:
		f := e.Evaluate(node.Function, env)
		if isError(f) {
			return f
		}
		var arg Value
		if node.Argument != nil {
			arg = &Thunk{expr: node.Argument, env: env}
		}
		return e.applyFunction(f, arg)
	}

	return error("Unknown node type: %T", node)
}

func (e *Evaluator) evalProgram(program *ast.Program, env *Environment) Value {
	var result Value

	for _, stmt := range program.Statements {
		result = e.Evaluate(stmt, env)
		if isError(result) {
			return result
		}
	}
	return result
}

func (e *Evaluator) evalBinding(node *ast.BindingStatement, env *Environment) Value {
	if env.Has(node.Name) {
		return error("'%s' is already defined", node.Name)
	}

	val := &Thunk{expr: node.Value, env: env, bindingName: node.Name}
	env.Set(node.Name, val)
	return val
}

func (e *Evaluator) evalIdentifier(node *ast.Identifier, env *Environment) Value {
	if node.Value == "nil" {
		return &Nil{}
	}

	if val, ok := env.Get(node.Value); ok {
		val = e.force(val)
		if isError(val) {
			return val
		}

		if fn, ok := val.(*Function); ok && (fn.Parameter == "" || fn.Parameter == "_") {
			return e.Evaluate(fn.Body, fn.Env)
		}
		return val
	}
	return &Undefined{}
}

func (e *Evaluator) evalPrefix(operator string, right Value) Value {
	right = e.force(right)
	switch operator {
	case "-":
		return e.evalMinusPrefix(right)
	}
	return error("Unknown operator: %s%s", operator, right.Type())
}

func (e *Evaluator) evalMinusPrefix(right Value) Value {
	switch v := right.(type) {
	case *Number:
		return &Number{Value: -v.Value}
	}

	return error("Unknown operator: -%s", right.Type())
}

func (e *Evaluator) evalInfix(operator string, left, right Value, env *Environment) Value {
	left = e.force(left)
	right = e.force(right)
	if isError(left) || isError(right) {
		return left
	}

	switch {
	case left.Type() == NumberValue && right.Type() == NumberValue:
		return e.evalNumberInfix(operator, left, right, env)

	case left.Type() == NilValue || right.Type() == NilValue:
		return e.evalNilInfix(operator, left, right, env)
	}
	return error("Invalid operation: %s %s %s", left.Type(), operator, right.Type())
}

func floatToInt(f float64) (int, bool) {
	if f == math.Trunc(f) {
		return int(f), true
	}
	return 0, false
}

func (e *Evaluator) evalBool(value bool, env *Environment) Value {
	var valStr string
	if value {
		valStr = "true"
	} else {
		valStr = "false"
	}
	return e.Evaluate(&ast.Identifier{
		Value: valStr,
	}, env)
}

func (e *Evaluator) evalNilInfix(operator string, left, right Value, env *Environment) Value {
	eq := true
	if operator == "!=" {
		eq = false
	}

	if _, ok := left.(*Nil); ok {
		if _, ok := right.(*Nil); ok {
			return e.evalBool(eq, env)
		}
	}
	return e.evalBool(false, env)
}

func (e *Evaluator) evalNumberInfix(operator string, left, right Value, env *Environment) Value {
	l := left.(*Number).Value
	r := right.(*Number).Value

	switch operator {
	case "+":
		return &Number{Value: l + r}
	case "-":
		return &Number{Value: l - r}
	case "*":
		return &Number{Value: l * r}
	case "/":
		if r == 0 {
			return error("Division by zero")
		}
		return &Number{Value: l / r}
	case "%":
		if r == 0 {
			return error("Modulo by zero")
		}
		if lInt, ok := floatToInt(l); ok {
			if rInt, ok := floatToInt(r); ok {
				return &Number{Value: float64(lInt % rInt)}
			}
		}
		return error("Modulo of float")

	case "=":
		return e.evalBool(l == r, env)
	case "!=":
		return e.evalBool(l != r, env)
	case ">":
		return e.evalBool(l > r, env)
	case ">=":
		return e.evalBool(l >= r, env)
	case "<":
		return e.evalBool(l < r, env)
	case "<=":
		return e.evalBool(l <= r, env)
	}

	return error("Invalid operation: %s %s %s", left.Type(), operator, right.Type())
}

func (e *Evaluator) applyFunction(f Value, arg Value) Value {
	f = e.force(f)
	if isError(f) {
		return f
	}

	fn, ok := f.(*Function)
	if !ok {
		return error("Not a function: %s", f.String())
	}

	if fn.Parameter == "" || fn.Parameter == "_" {
		if arg != nil {
			return error("Function expects no arguments")

		}
		return e.Evaluate(fn.Body, fn.Env)
	}

	if arg == nil {
		return error("Function expects argument")
	}

	childEnv := NewEnvIn(fn.Env)
	childEnv.Set(fn.Parameter, arg)
	return e.Evaluate(fn.Body, childEnv)
}

func (e *Evaluator) force(v Value) Value {
	if thunk, ok := v.(*Thunk); ok {
		return thunk.Force(e)
	}
	return v
}