package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"time"
	"mylang/ast"
	"mylang/compiler"
	"mylang/evaluator"
	"mylang/lexer"
	"mylang/object"
	"mylang/parser"
	"mylang/repl"
	"mylang/vm"
)

var engine *string = flag.String("engine", "vm", "use 'vm' or 'eval'")
var input *string = flag.String("file", "repl", "use filename")
var benchmark *string = flag.String("bench", "no", "use 'yes' or 'no'")

func main() {
	flag.Parse()

	var inputFile []byte
	var result object.Object
	var duration time.Duration

	if *input == "repl" {
		user, err := user.Current()
		if err != nil {
			panic(err)
		}

		fmt.Printf("Hello %s, this is mylang\n", user.Username)
		repl.Start(os.Stdin, os.Stdout)
		
		return
	} else {
		var err error
		inputFile, err = os.ReadFile(*input)
		if err != nil {
			fmt.Printf("could not read: %s\n", err.Error())
		}
	}

	l := lexer.New(string(inputFile))
	p := parser.New(l)
	var program *ast.Program = p.ParseProgram()
	if len(p.Errors()) > 0 {
		repl.PrintParseErrors(os.Stdout, p.Errors())	
	}

	if *engine == "vm" {
		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			fmt.Printf("compile error: %s", err)
			return
		}

		machine := vm.New(comp.MakeBytecode())
		start := time.Now()

		err = machine.Run()
		if err != nil {
			fmt.Printf("virtual machine error: %s", err)
			return
		}

		duration = time.Since(start)
		result = machine.LastPoppedStackElement()
	} else {
		env := object.NewEnvironment()
		start := time.Now()
		result = evaluator.Evaluate(program, env)
		duration = time.Since(start)
	}

	fmt.Printf(
		"engine=%s, result=%s\n",
		*engine,
		result.Inspect(),
	)

	if *benchmark == "yes" {
		fmt.Printf("duration=%s\n", duration)
	}
}
