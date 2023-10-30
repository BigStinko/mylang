package lexer

import ( 
	"testing"
	"mylang/token"
)

// tests the tokens created by the lexer for the given input string
// against an array of the correct tokens and prints any errors
func TestNextToken(t *testing.T) {
	var input string = `let five = 5;
	let ten = 10;

	let add = func(x, y) {
		x + y;
	};

	let result = add(five, ten);
	!-/*5;
	5 < 10 > 5;

	if (5 < 10) {
		return true;
	} else {
		return false;
	}

	10 == 10;
	10 != 9;
	`

	tests := []struct {
		Type token.TokenType
		Literal string
	} {
		{token.LET, "let"},
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SCOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "ten"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SCOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.FUNCTION, "func"},
		{token.OPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.CPAREN, ")"},
		{token.OBRACE, "{"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SCOLON, ";"},
		{token.CBRACE, "}"},
		{token.SCOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.OPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.CPAREN, ")"},
		{token.SCOLON, ";"},
		{token.BANG, "!"},
		{token.MINUS, "-"},
		{token.SLASH, "/"},
		{token.ASTERISK, "*"},
		{token.INT, "5"},
		{token.SCOLON, ";"},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.GT, ">"},
		{token.INT, "5"},
		{token.SCOLON, ";"},
		{token.IF, "if"},
		{token.OPAREN, "("},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.CPAREN, ")"},
		{token.OBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SCOLON, ";"},
		{token.CBRACE, "}"},
		{token.ELSE, "else"},
		{token.OBRACE, "{"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SCOLON, ";"},
		{token.CBRACE, "}"},
		{token.INT, "10"},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.SCOLON, ";"},
		{token.INT, "10"},
		{token.NOT_EQ, "!="},
		{token.INT, "9"},
		{token.SCOLON, ";"},
		{token.EOF, ""},
	}

	var l *Lexer = New(input)

	for i, testToken := range tests {
		var tok token.Token = l.NextToken()

		if tok.Type != testToken.Type {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q", i, testToken.Type, tok.Type)
		}
		
		if tok.Literal != testToken.Literal {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q", i, testToken.Literal, tok.Literal)
		}
	}
}
