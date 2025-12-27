package main

import (
	"flag"
	"lambda/src/interpreter"
	"os"
)

func main() {
	fileFlag := flag.String("f", "", "Execute a file")
	flag.Parse()

	if *fileFlag != "" {
		interpreter.ExecuteFile(*fileFlag, os.Stdout)
		os.Exit(0)
	}

	interpreter.StartREPL(os.Stdin, os.Stdout)
}