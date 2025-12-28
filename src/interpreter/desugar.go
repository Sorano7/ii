package interpreter

import "lambda/src/ast"

func desugarProgram(node *ast.Program) *ast.Program {
	stmt := make([]ast.Statement, 0)
	for _, s := range node.Statements {
		desugared := desugarStatement(s)
		stmt = append(stmt, desugared)
	}
	
	return &ast.Program{Statements: stmt}
}

func desugarStatement(node ast.Statement) ast.Statement {
	switch n := node.(type) {
	case *ast.BindingStatement:
		return &ast.BindingStatement{
			Name: n.Name,
			Value: desugarExpression(n.Value),
		}

	case *ast.ExpressionStatement:
		return &ast.ExpressionStatement{
			Expression: desugarExpression(n.Expression),
		}
	}
	return node
}

func desugarExpression(node ast.Expression) ast.Expression {
	switch n := node.(type) {
	case *ast.PairLiteral:
		return &ast.CallExpression{
			Function: &ast.CallExpression{
				Function: &ast.Identifier{Value: "pair"},
				Argument: desugarExpression(n.First),
			},
			Argument: desugarExpression(n.Second),
		}

	case *ast.CallExpression:
		return &ast.CallExpression{
			Function: desugarExpression(n.Function),
			Argument: desugarExpression(n.Argument),
		}

	case *ast.FunctionLiteral:
		return &ast.FunctionLiteral{
			Parameter: n.Parameter,
			Body: desugarExpression(n.Body),
		}

	case *ast.InfixExpression:
		switch left := n.Left.(type) {
		case *ast.StringLiteral:
			n.Left = stringToArray(left)
		}
		switch right := n.Right.(type) {
		case *ast.StringLiteral:
			n.Right = stringToArray(right)
		}

		if leftArr, ok := n.Left.(*ast.ArrayLiteral); ok && n.Operator == "+" {
			if rightArr, ok := n.Right.(*ast.ArrayLiteral); ok {
				return &ast.CallExpression{
					Function: &ast.CallExpression{
						Function: &ast.Identifier{Value: "concat"},
						Argument: desugarExpression(leftArr),
					},
					Argument: desugarExpression(rightArr),
				}
			}
		}

		return &ast.InfixExpression{
			Left: desugarExpression(n.Left),
			Operator: n.Operator,
			Right: desugarExpression(n.Right),
		}

	case *ast.PrefixExpression:
		return &ast.PrefixExpression{
			Operator: n.Operator,
			Right: desugarExpression(n.Right),
		}

	case *ast.PipeExpression:
		switch n.Direction {
		case "<<", "<>", ".":
			return &ast.CallExpression{
				Function: desugarExpression(n.Left),
				Argument: desugarExpression(n.Right),
			}
		case ">>":
			return &ast.CallExpression{
				Function: desugarExpression(n.Right),
				Argument: desugarExpression(n.Left),
			}

		default:
			panic("Unknown pipe direction")
		}

	case *ast.ArrayLiteral:
		result := ast.Expression(&ast.Identifier{Value: "nil"})
		for i := len(n.Elements) - 1; i >= 0; i-- {
			result = &ast.CallExpression{
				Function: &ast.CallExpression{
					Function: &ast.Identifier{Value: "cons"},
					Argument: desugarExpression(n.Elements[i]),
				},
				Argument: result,
			}
		}
		return result

	case *ast.StringLiteral:
		return desugarExpression(stringToArray(n))

	case *ast.CondExpression:
		return &ast.CallExpression{
			Function: &ast.CallExpression{
				Function: desugarExpression(n.If),
				Argument: desugarExpression(n.Then),
			},
			Argument: desugarExpression(n.Else),
		}
	}

	return node
}

func stringToArray(node *ast.StringLiteral) *ast.ArrayLiteral {
	chars := make([]ast.Expression, 0)
	for _, c := range node.Value {
		char := &ast.CharLiteral{Value: "'" + string(c) + "'"}
		if char.Value != "'\"'" {
			chars = append(chars, char)
		}
	}
	return &ast.ArrayLiteral{Elements: chars}
}