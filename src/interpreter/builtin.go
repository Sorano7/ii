package interpreter

import "fmt"

var builtins = map[string]*Builtin{
	"write": { Name: "write", Fn: builtinWrite },
}

func builtinWrite(arg Value) Value {
	if arg != nil {
		fmt.Println(arg.String())
	} else {
		fmt.Println()
	}
	return &Number{Value: 0}
}
