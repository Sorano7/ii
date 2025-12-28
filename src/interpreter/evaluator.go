package interpreter

import (
	"lambda/src/ast"
	"math"
)

type Evaluator struct{}

func createEvaluator() *Evaluator {
	return &Evaluator{}
}

func (e *Evaluator) EvaluateProgram(program *ast.Program, env *Environment, exprCount int) Value {
	count := 0

	var result Value

	for _, stmt := range program.Statements {
		if _, ok := stmt.(*ast.ExpressionStatement); ok {
			if count >= exprCount {
				return error("Too many expressions")
			}
			count++
		}
		result = e.Evaluate(stmt, env)
		if isError(result) {
			return result
		}
	}
	return result
}

func (e *Evaluator) Evaluate(node ast.Node, env *Environment) Value {
	switch node := node.(type) {
	case *ast.BindingStatement:
		return e.evalBinding(node, env)

	case *ast.ExpressionStatement:
		return e.Evaluate(node.Expression, env)

	case *ast.NumberLiteral:
		return &Number{Value: node.Value}

	case *ast.CharLiteral:
		return &Char{Value: node.Value}

	case *ast.PairLiteral:
		return &Pair{
			First:  &Thunk{expr: node.First, env: env},
			Second: &Thunk{expr: node.Second, env: env},
		}

	case *ast.ArrayLiteral:
		thunks := make([]Value, 0)
		for _, el := range node.Elements {
			thunks = append(thunks, &Thunk{expr: el, env: env})
		}
		return &Array{
			Elements: thunks,
		}

	case *ast.StringLiteral:
		return &String{
			Value: node.Value,
		}

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
		return e.evalInfix(node.Operator, left, right)

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

func (e *Evaluator) evalBinding(node *ast.BindingStatement, env *Environment) Value {
	if env.Has(node.Name) {
		return error("'%s' is already defined", node.Name)
	}

	val := &Thunk{expr: node.Value, env: env, bindingName: node.Name}
	env.Set(node.Name, val)
	return val
}

func (e *Evaluator) evalIdentifier(node *ast.Identifier, env *Environment) Value {
	switch node.Value {
	case "nil":
		return &Nil{}
	case "true", "false":
		return makeBool(node.Value == "true")
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

	if builtin, ok := builtins[node.Value]; ok {
		if builtin.Name == "read" {
			return e.applyFunction(builtin, nil)
		}
		return builtin
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

func (e *Evaluator) evalInfix(operator string, left, right Value) Value {
	left = e.force(left)
	right = e.force(right)
	if isError(left) || isError(right) {
		return left
	}

	switch operator {
	case "+":
		return e.evalPlus(left, right)
	case "-":
		return e.evalDash(left, right)
	case "*":
		return e.evalStar(left, right)
	case "/":
		return e.evalSlash(left, right)
	case "%":
		return e.evalPercent(left, right)
	case "=", "!=":
		return makeBool(e.evalEquals(operator == "=", left, right))
	case "<", "<=", ">", ">=":
		switch {
		case left.Type() == NumberValue && right.Type() == NumberValue:
			return makeBool(e.compareNumber(operator, left, right))
		case left.Type() == CharValue && right.Type() == CharValue:
			return makeBool(e.compareChar(operator, left, right))
		}
		return error("Invalid comparison: %s %s %s", left.Type(), operator, right.Type())
	}

	return error("Unknown operator: %s", operator)
}

func (e *Evaluator) evalPlus(left, right Value) Value {
	left = e.force(left)
	right = e.force(right)

	switch {
	case left.Type() == NumberValue && right.Type() == NumberValue:
		return &Number{Value: left.(*Number).Value + right.(*Number).Value}

	case left.Type() == CharValue && right.Type() == CharValue:
		return &String{Value: left.(*Char).Value + right.(*Char).Value}

	case left.Type() == StringValue && right.Type() == StringValue:
		return &String{Value: left.(*String).Value + right.(*String).Value}

	case left.Type() == ArrayValue && right.Type() == ArrayValue:
		return &Array{
			Elements: append(left.(*Array).Elements, right.(*Array).Elements...),
		}
	}
	return error("Invalid operation: %s + %s", left.Type(), right.Type())
}

func (e *Evaluator) evalDash(left, right Value) Value {
	left = e.force(left)
	right = e.force(right)

	switch {
	case left.Type() == NumberValue && right.Type() == NumberValue:
		return &Number{Value: left.(*Number).Value - right.(*Number).Value}
	}
	return error("Invalid operation: %s - %s", left.Type(), right.Type())
}

func (e *Evaluator) evalStar(left, right Value) Value {
	left = e.force(left)
	right = e.force(right)

	switch {
	case left.Type() == NumberValue && right.Type() == NumberValue:
		return &Number{Value: left.(*Number).Value * right.(*Number).Value}
	}
	return error("Invalid operation: %s * %s", left.Type(), right.Type())
}

func (e *Evaluator) evalSlash(left, right Value) Value {
	left = e.force(left)
	right = e.force(right)

	switch {
	case left.Type() == NumberValue && right.Type() == NumberValue:
		if right.(*Number).Value == 0 {
			return error("Division by zero")
		}
		return &Number{Value: left.(*Number).Value / right.(*Number).Value}
	}
	return error("Invalid operation: %s / %s", left.Type(), right.Type())
}

func (e *Evaluator) evalPercent(left, right Value) Value {
	left = e.force(left)
	right = e.force(right)

	switch {
	case left.Type() == NumberValue && right.Type() == NumberValue:
		l, lok := floatToInt(left.(*Number).Value)
		r, rok := floatToInt(right.(*Number).Value)
		if !lok || !rok {
			return error("Modulo of float")
		}
		if r == 0 {
			return error("Modulo by zezo")
		}
		return &Number{Value: float64(l % r)}
	}
	return error("Invalid operation: %s - %s", left.Type(), right.Type())
}

func (e *Evaluator) evalEquals(expectsEqual bool, left, right Value) bool {
	left = e.force(left)
	right = e.force(right)

	expr := false
	switch {
	case left.Type() == NumberValue && right.Type() == NumberValue:
		expr = left.(*Number).Value == right.(*Number).Value

	case left.Type() == CharValue && right.Type() == CharValue:
		expr = left.(*Char).Value == right.(*Char).Value

	case left.Type() == StringValue && right.Type() == StringValue:
		expr = left.(*String).Value == right.(*String).Value

	case left.Type() == PairValue && right.Type() == PairValue:
		expr = e.evalPairEquals(left, right)

	case left.Type() == ArrayValue && right.Type() == ArrayValue:
		expr = e.evalArrayEquals(left, right)

	case left.Type() == BoolValue && right.Type() == BoolValue:
		expr = left.(*Bool).Value == right.(*Bool).Value

	case left.Type() == NilValue && right.Type() == NilValue:
		expr = true

	case left.Type() == StringValue && right.Type() == NilValue:
		expr = left.(*String).Value == ""
	case left.Type() == ArrayValue && right.Type() == NilValue:
		expr = len(left.(*Array).Elements) == 0
	}

	return expr == expectsEqual
}

func (e *Evaluator) evalPairEquals(left, right Value) bool {
	l := left.(*Pair)
	r := right.(*Pair)

	lf := e.force(l.First)
	if isError(lf) {
		return false
	}
	rf := e.force(r.First)
	if isError(rf) {
		return false
	}

	ls := e.force(l.Second)
	if isError(ls) {
		return false
	}
	rs := e.force(r.Second)
	if isError(rs) {
		return false
	}

	return e.evalEquals(true, l.First, r.First) && e.evalEquals(true, l.Second, r.Second)
}

func (e *Evaluator) evalArrayEquals(arr1, arr2 Value) bool {
	a1 := arr1.(*Array)
	a2 := arr2.(*Array)

	if len(a1.Elements) != len(a2.Elements) {
		return false
	}

	for i := range a1.Elements {
		left := e.force(a1.Elements[i])
		if isError(left) {
			return false
		}
		right := e.force(a2.Elements[i])
		if isError(right) {
			return false
		}
		if e.evalEquals(false, left, right) {
			return false
		}
	}
	return true
}

func (e *Evaluator) compareNumber(operator string, left, right Value) bool {
	l := left.(*Number)
	r := right.(*Number)
	switch operator {
	case ">":
		return l.Value > r.Value
	case ">=":
		return l.Value >= r.Value
	case "<":
		return l.Value < r.Value
	case "<=":
		return l.Value <= r.Value
	}
	return false
}

func (e *Evaluator) compareChar(operator string, left, right Value) bool {
	l := left.(*Char)
	r := right.(*Char)
	switch operator {
	case ">":
		return l.Value > r.Value
	case ">=":
		return l.Value >= r.Value
	case "<":
		return l.Value < r.Value
	case "<=":
		return l.Value <= r.Value
	}
	return false
}

func makeBool(value bool) Value {
	return &Bool{Value: value}
}

func floatToInt(f float64) (int, bool) {
	if f == math.Trunc(f) {
		return int(f), true
	}
	return 0, false
}

func (e *Evaluator) applyFunction(f Value, arg Value) Value {
	f = e.force(f)
	if isError(f) {
		return f
	}

	switch fn := f.(type) {
	case *Function:
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

	case *Pair:
		result := e.applyFunction(arg, fn.First)
		return e.applyFunction(result, fn.Second)

	case *Array:
		return e.applyFunction(arrayToPair(fn), arg)

	case *String:
		return e.applyFunction(stringToPair(fn), arg)

	case *Builtin:
		if arg != nil && arg.Type() == ThunkValue {
			return e.applyFunction(fn, e.force(arg))
		}
		return fn.Fn(e.force(arg))

	case *Bool:
		return &BoolThunk{Cond: fn.Value, First: arg}

	case *BoolThunk:
		if fn.Cond {
			return fn.First
		}
		return arg

	case *Thunk:
		return e.applyFunction(e.force(fn), arg)

	default:
		return error("Not a function: %s", fn.Type())
	}
}

func (e *Evaluator) force(v Value) Value {
	if thunk, ok := v.(*Thunk); ok {
		return thunk.Force(e)
	}
	return v
}

func stringToPair(str *String) Value {
	if str.Value == "" {
		return &Nil{}
	}
	
	chars := make([]Value, 0)
	for _, s := range str.Value {
		chars = append(chars, &Char{Value: string(s)})
	}

	return arrayToPair(&Array{Elements: chars})
}

func arrayToPair(arr *Array) Value {
	if len(arr.Elements) == 0 {
		return &Nil{}
	}

	result := Value(&Nil{})
	for i := len(arr.Elements) - 1; i >= 0; i-- {
		result = &Pair{
			First:  arr.Elements[i],
			Second: result,
		}
	}
	return result.(*Pair)
}
