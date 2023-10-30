// Read-Eval-Print-Loop
package repl

import (
	"bufio"
	"fmt"
	"io"
	"mylang/lexer"
	"mylang/token"
)

const PROMPT = ">> "

// takes an input and an output, reads the text from the input
// evaluates the input in the lexer, and prints the tokens to
// the out
func StartREPL(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Fprintf(out, PROMPT)
		var scanned bool = scanner.Scan()
		if !scanned {
			return
		}

		var line string = scanner.Text()
		var l *lexer.Lexer = lexer.New(line)

		for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
			fmt.Fprintf(out, "%+v\n", tok)
		}
	}
}
