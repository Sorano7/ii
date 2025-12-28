package interpreter

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

var builtins = map[string]*Builtin{
	"write": {Name: "write", Fn: builtinWrite},
	"read":  {Name: "read", Fn: builtinRead},
	"string":   {Name: "string", Fn: builtinString},
	"number":   {Name: "number", Fn: builtinNumber},
	"chr":   {Name: "chr", Fn: builtinChr},
	"ord":   {Name: "ord", Fn: builtinOrd},
}

func builtinWrite(arg Value) Value {
	if arg == nil {
		fmt.Println()
		return arg
	}

	switch v := arg.(type) {
	case *String:
		fmt.Println(v.Value)
	case *Char:
		fmt.Println(v.Value)
	default:
		fmt.Println(v.String())
	}
	return arg
}

func builtinRead(arg Value) Value {
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

func builtinString(arg Value) Value {
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

func builtinNumber(arg Value) Value {
	if isNil(arg) {
		return error("Invalid argument")
	}
	switch v := arg.(type) {
	case *Number:
		return v

	case *String:
		val, err := strconv.ParseFloat(v.Value, 64)
		if err != nil {
			return error("Cannot parse as number: %s", v.Value)
		}
		return &Number{Value: val}

	default:
		return error("Invalid argument type: %s", arg.Type())
	}
}

func builtinOrd(arg Value) Value {
	switch v := arg.(type) {
	case *Char:
		chars := []rune(v.Value)
		return &Number{Value: float64(int(chars[0]))}
	default:
		return error("Invalid argument type: %s", arg.Type())
	}
}

func builtinChr(arg Value) Value {
	switch v := arg.(type) {
	case *Number:
		i, ok := floatToInt(v.Value)
		if !ok {
			return error("Invalid argument type: float")
		}
		return &Char{Value: string(rune(i))}

	default:
		return error("Invalid argument type: %s", arg.Type())
	}
}

func isNil(arg Value) bool {
	return arg == nil || arg.Type() == NilValue
}