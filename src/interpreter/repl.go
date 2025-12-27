package interpreter

import (
	"bufio"
	"fmt"
	"io"
	"lambda/src/parser"
	"os"
	"strings"
)

const Prompt = "|> "

type REPL struct {
	env    *Environment
	eval   *Evaluator
	out    io.Writer
	buffer strings.Builder
}

func createRepl(out io.Writer) *REPL {
	repl := &REPL{
		env: NewEnv(),
		eval: createEvaluator(),
		out: out,
	}
	repl.LoadFile("lib/prelude.x", false, false)
	return repl
}

func (r *REPL) evaluateAndPrint(input string) {
	result := r.evaluate(input)
	if result == nil {
		return		
	}
	result = r.eval.force(result)
	if result == nil {
		return
	}
	fmt.Fprintln(r.out, result.String())
}

func (r *REPL) evaluate(src string) Value {
	program, ok := parser.Parse(src)
	if !ok {
		fmt.Fprintln(r.out, "[Error] Invalid syntax")
		return nil
	}
	program = desugarProgram(program)
	return r.eval.Evaluate(program, r.env)
}

func (r *REPL) handleCommand(input string) bool {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return true
	}

	cmd := parts[0]

	switch cmd {
	case ":q", ":quit":
		return false

	case ":c", ":clear":
		r.env = NewEnv()
		fmt.Fprintln(r.out, "Environment cleared.")
		return true

	case ":l", ":load":
		if len(parts) > 1 {
			r.LoadFile(parts[1], false, false)
		}
		return true

	case ":e", ":exec":
		if len(parts) > 1 {
			r.LoadFile(parts[1], true, true)
		}
		return true

	case ":a", ":ast":
		if len(parts) > 1 {
			r.PrintAST(strings.Join(parts[1:], ""))
		}
		return true
	}
	fmt.Fprintf(r.out, "Unknown command: %s\n", cmd)
	return true
}

func (r *REPL) PrintAST(src string) {
	program, ok := parser.Parse(src)
	if !ok {
		fmt.Fprintln(r.out, "[Error] Invalid syntax")
		return
	}
	program = desugarProgram(program)
	fmt.Fprint(r.out, program.Debug())
}

func (r *REPL) LoadFile(filename string, clearEnv bool, printMain bool) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(r.out, "[Error] %s\n", err)
		return
	}
	_ = r.evaluate(string(bytes))

	if value, ok := r.env.Get("main"); ok {
		value = r.eval.force(value)

		if printMain || isError(value) {
			fmt.Fprintln(r.out, value.String())
		}
	}

	if clearEnv {
		r.env = NewEnv()
	}
}

func StartREPL(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	repl := createRepl(out)

	for {
		fmt.Fprint(repl.out, Prompt)

		if !scanner.Scan() {
			break
		}

		input := scanner.Text()

		if strings.HasPrefix(input, ":") {
			if repl.handleCommand(input) {
				continue
			}
			break
		}

		if strings.TrimSpace(input) == "" {
			continue
		}

		repl.evaluateAndPrint(input)
	}
}

func ExecuteFile(filename string, out io.Writer) {
	repl := createRepl(out)
	repl.LoadFile(filename, false, true)
}