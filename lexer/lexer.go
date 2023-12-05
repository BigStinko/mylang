package lexer

import (
	"mylang/token"
	"unicode"
)

type Lexer struct {
	input []rune
	position int
	readPosition int
	char rune
}

func New(input string) *Lexer {
	l := &Lexer{input: []rune(input)}
	l.readChar()
	return l
}

// creates the next token in the lexer input
func (l *Lexer) NextToken() (tok token.Token) {
	l.skipWhitespace()

	// skip single line comment
	if l.char == rune('/') && l.peekChar() == rune('/') {
		for l.char != '\n' && l.char != rune(0) {
			l.readChar()
		}
		l.skipWhitespace()

		return l.NextToken()
	}

	// skipe block comment
	if l.char == rune('/') && l.peekChar() == rune('*') {
		var end bool = false

		for !end {
			if l.char == rune(0) {
				end = true
			}

			if l.char == '*' && l.peekChar() == '/' {
				end = true

				l.readChar()
			}

			l.readChar()
		}

		l.skipWhitespace()
	}

	switch l.char {
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: "=="}
		} else {
			tok = token.Token{Type: token.ASSIGN, Literal: string(l.char)}
		}
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
	case ';':
		tok = token.Token{Type: token.SCOLON, Literal: string(l.char)}
	case ':':
		tok = token.Token{Type: token.COLON, Literal: string(l.char)}
	case '(':
		tok = token.Token{Type: token.OPAREN, Literal: string(l.char)}
	case ')':
		tok = token.Token{Type: token.CPAREN, Literal: string(l.char)}
	case '{':
		tok = token.Token{Type: token.OBRACE, Literal: string(l.char)}
	case '}':
		tok = token.Token{Type: token.CBRACE, Literal: string(l.char)}
	case '[':
		tok = token.Token{Type: token.OBRACKET, Literal: string(l.char)}
	case ']':
		tok = token.Token{Type: token.CBRACKET, Literal: string(l.char)}
	case ',':
		tok = token.Token{Type: token.COMMA, Literal: string(l.char)}
	case '+':
		tok = token.Token{Type: token.PLUS, Literal: string(l.char)}
	case '-':
		tok = token.Token{Type: token.MINUS, Literal: string(l.char)}
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: "!="}
		} else {
			tok = token.Token{Type: token.BANG, Literal: string(l.char)}
		}
	case '/':
		tok = token.Token{Type: token.SLASH, Literal: string(l.char)}
	case '*':
		tok = token.Token{Type: token.ASTERISK, Literal: string(l.char)}
	case '%':
		tok = token.Token{Type: token.MODULO, Literal: string(l.char)}
	case '<':
		tok = token.Token{Type: token.LT, Literal: string(l.char)}
	case '>':
		tok = token.Token{Type: token.GT, Literal: string(l.char)}

	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isDigit(l.char) {
			tok = l.readNumber()
			return tok
		}

		if !unicode.IsPrint(l.char) {
			tok.Literal = string(l.char)
			tok.Type = token.ILLEGAL

			l.readChar()
			return tok
		}

		tok.Literal = l.readIdentifier()

		if len(tok.Literal) == 0 {
			tok.Type = token.ILLEGAL
			l.readChar()
			return tok
		}

		tok.Type = token.LookupIdent(tok.Literal)
		return tok
	}

	l.readChar()   // go to the next character
	return tok
}

func (l *Lexer) GetLine() int {
	var line int = 0
	var length int = len(l.input)
	
	for i := 0; i < l.readPosition && i < length; i++ {
		if l.input[i] == rune('\n') {
			line++
		}
	}

	return line
}

func isLetter(char rune) bool {
	return 'a' <= char && char <= 'z' || 'A' <= char && char <= 'Z' || char == '_'
}

func isDigit(char rune) bool {
	return '0' <= char && char <= '9'
}

// reads a character from the lexer's input string and increments
// the position and readPosition values so that position refers
// to the current character being read by the lexer and readPosition
// refers to the next character being read. If readPositon is 
// advanced past the end of the string, the character is set to
// the eof character
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.char = 0
	} else {
		l.char = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) readNumber() token.Token {
	var number string = l.readDigits()

	if l.char == '.' && isDigit(l.peekChar()) {
		l.readChar()
		var decimal string = l.readDigits()
		return token.Token{Type: token.FLOAT, Literal: number + "." + decimal}
	}
	return token.Token{Type: token.INT, Literal: number}
}

// when the lexer gets to a number in the input this function
// reads through the number in the string and returns it.
func (l *Lexer) readDigits() string {
	var startPos int = l.position
	for isDigit(l.char) {
		l.readChar()
	}
	return string(l.input[startPos:l.position])
}

// extracts a string from the lexer surrounded by quotations
func (l *Lexer) readString() string {
	var out string = ""

	for {
		l.readChar()
		if l.char == '"' || l.char == 0 {
			break
		}

		if l.char == '\\' {
			l.readChar()

			switch l.char {
			case 'n':
				l.char = '\n'
			case 'r':
				l.char = '\r'
			case 't':
				l.char = '\t'
			case '"':
				l.char = '"'
			case '\\':
				l.char = '\\'
			}
		}
		
		out = out + string(l.char)
	}

	return out
}

// reads through any identifier/keyword in the input string and
// returns it 
func (l *Lexer) readIdentifier() string {
	var startPos int = l.position
	for isLetter(l.char) || isDigit(l.char) {
		l.readChar()
	}
	return string(l.input[startPos:l.position])
}

// advances the lexer through the input string until it finds a
// character that isn't whitespace
func (l *Lexer) skipWhitespace() {
	for l.char == ' ' || l.char == '\n' || l.char == '\t' || l.char == '\r' {
		l.readChar()
	}
}

// looks at the next character 
func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0;
	} else {
		return l.input[l.readPosition]
	}
}
