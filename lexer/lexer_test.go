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
	5 < 10 > 5;

	if (5 < 10) {
		return true;
	} else {
		return false;
	}

	10 == 10;
	10 != 9;
	"foo\n\t\r\"\\bar"
	"foo bar"
	[1, 2];
	{"foo": "bar"}
	true and false;
	false or true;
	10 % 9;
	'a';
	0.333;
	while (x > y) {
		let x = x - 1;
	}
	
	switch (x) {
	case true {
		a
	}
	case false {
		b
	}
	default {
		c
	}
	}
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
		{token.STRING, "foo\n\t\r\"\\bar"},
		{token.STRING, "foo bar"},
		{token.OBRACKET, "["},
		{token.INT, "1"},
		{token.COMMA, ","},
		{token.INT, "2"},
		{token.CBRACKET, "]"},
		{token.SCOLON, ";"},
		{token.OBRACE, "{"},
		{token.STRING, "foo"},
		{token.COLON, ":"},
		{token.STRING, "bar"},
		{token.CBRACE, "}"},
		{token.TRUE, "true"},
		{token.AND, "and"},
		{token.FALSE, "false"},
		{token.SCOLON, ";"},
		{token.FALSE, "false"},
		{token.OR, "or"},
		{token.TRUE, "true"},
		{token.SCOLON, ";"},
		{token.INT, "10"},
		{token.MODULO, "%"},
		{token.INT, "9"},
		{token.SCOLON, ";"},
		{token.BYTE, "a"},
		{token.SCOLON, ";"},
		{token.FLOAT, "0.333"},
		{token.SCOLON, ";"},
		{token.WHILE, "while"},
		{token.OPAREN, "("},
		{token.IDENT, "x"},
		{token.GT, ">"},
		{token.IDENT, "y"},
		{token.CPAREN, ")"},
		{token.OBRACE, "{"},
		{token.LET, "let"},
		{token.IDENT, "x"},
		{token.ASSIGN, "="},
		{token.IDENT, "x"},
		{token.MINUS, "-"},
		{token.INT, "1"},
		{token.SCOLON, ";"},
		{token.CBRACE, "}"},
		{token.SWITCH, "switch"},
		{token.OPAREN, "("},
		{token.IDENT, "x"},
		{token.CPAREN, ")"},
		{token.OBRACE, "{"},
		{token.CASE, "case"},
		{token.TRUE, "true"},
		{token.OBRACE, "{"},
		{token.IDENT, "a"},
		{token.CBRACE, "}"},
		{token.CASE, "case"},
		{token.FALSE, "false"},
		{token.OBRACE, "{"},
		{token.IDENT, "b"},
		{token.CBRACE, "}"},
		{token.DEFAULT, "default"},
		{token.OBRACE, "{"},
		{token.IDENT, "c"},
		{token.CBRACE, "}"},
		{token.CBRACE, "}"},
		{token.EOF, ""},
	}

	var l *Lexer = New(input)

	for i, test := range tests {
		var tok token.Token = l.NextToken()

		if tok.Type != test.Type {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, test.Type, tok.Type)
		}
		
		if tok.Literal != test.Literal {
			t.Fatalf("tests[%d] - literal wrong. expected=%q len=%d, got=%q len=%d",
				i, test.Literal, len(test.Literal), tok.Literal, len(tok.Literal))
		}
	}
}
