package interpreter

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

func (e *Evaluator) print(arg Value) Value {
	if arg == nil {
		fmt.Println()
		return &Nil{}
	}
	fmt.Println(arg.String())
	return &Number{Value: 0}
}

func (e *Evaluator) read(arg Value) Value {
	if !isNil(arg) {
		return error("Function expects no argument")
	}

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		input := scanner.Text()
		return &String{Value: input}
	}
	if err := scanner.Err(); err != nil {
		return error("%s", err)
	}
	return &String{Value: ""}
}

func (e *Evaluator) stringify(arg Value) Value {
	if isNil(arg) {
		return &String{Value: ""}
	}
	switch v := arg.(type) {
	case *String:
		return v
	default:
		return &String{Value: v.String()}
	}
}

func (e *Evaluator) number(arg Value) Value {
	if isNil(arg) {
		return error("Invalid argument")
	}
	switch v := arg.(type) {
	case *Number:
		return v

	case *String:
		val, err := strconv.ParseFloat(v.Value, 64)
		if err != nil {
			return &Nil{}
		}
		return &Number{Value: val}

	case *Char:
		val, err := strconv.ParseFloat(v.Value, 64)
		if err != nil {
			return &Nil{}
		}
		return &Number{Value: val}

	default:
		return &Nil{}
	}
}

func (e *Evaluator) ord(arg Value) Value {
	switch v := arg.(type) {
	case *Char:
		chars := []rune(v.Value)
		return &Number{Value: float64(int(chars[0]))}
	default:
		return &Nil{}
	}
}

func (e *Evaluator) chr(arg Value) Value {
	switch v := arg.(type) {
	case *Number:
		i, ok := floatToInt(v.Value)
		if !ok {
			return &Nil{}
		}
		return &Char{Value: string(rune(i))}

	default:
		return &Nil{}
	}
}

func isNil(arg Value) bool {
	return arg == nil || arg.Type() == NilValue
}