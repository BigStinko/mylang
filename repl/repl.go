// Read-Eval-Print-Loop
package repl

import (
	"bufio"
	"fmt"
	"io"
	"mylang/ast"
	"mylang/evaluator"
	"mylang/lexer"
	"mylang/object"
	"mylang/parser"
)

const PROMPT = ">> "

func printParseErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		io.WriteString(out, "\t" + msg + "\n")
	}
}

// takes an input and an output, reads the text from the input
// evaluates the input in the lexer, and prints the tokens to
// the out
func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Fprintf(out, PROMPT)
		var scanned bool = scanner.Scan()
		if !scanned {
			return
		}

		var line string = scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)
		var program *ast.Program = p.ParseProgram()
		
		if len(p.Errors()) != 0 {
			printParseErrors(out, p.Errors())
			continue
		}

		var evaluated object.Object = evaluator.Evaluate(program)
		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
	}
}
