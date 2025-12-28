package main

import (
	"flag"
	"lambda/src/interpreter"
	"os"
)

func main() {
	fileFlag := flag.String("f", "", "Execute a file")
	printFlag := flag.Bool("o", false, "Output the main expression")
	flag.Parse()

	if *fileFlag != "" {
		interpreter.ExecuteFile(*fileFlag, os.Stdout, *printFlag)
		os.Exit(0)
	}

	interpreter.StartREPL(os.Stdin, os.Stdout)
}