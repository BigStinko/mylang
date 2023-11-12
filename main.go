package main

import (
	"fmt"
	"mylang/ast"
	"mylang/evaluator"
	"mylang/lexer"
	"mylang/object"
	"mylang/parser"
	"mylang/repl"
	"os"
	"os/user"
)
func main() {

	env := object.NewEnvironment()

	if len(os.Args) > 1 {
		input, err := os.ReadFile(os.Args[1])
		
		if err != nil {
			fmt.Printf("could not read: %s\n", err.Error())
		} else {
			l := lexer.New(string(input))
			p := parser.New(l)
			var program *ast.Program = p.ParseProgram()
			if len(p.Errors()) > 0 {
				repl.PrintParseErrors(os.Stdout, p.Errors())	
			}
			var evaluated object.Object = evaluator.Evaluate(program, env)

			if evaluated != nil && evaluated != evaluator.NULL{
				fmt.Print(evaluated.Inspect())
			}
		}

		return
	}

	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Hello %s, this is mylang\n", user.Username)
	repl.Start(os.Stdin, os.Stdout)
}
