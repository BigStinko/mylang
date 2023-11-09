package token

type TokenType string

type Token struct {
	Type TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF = "EOF"
	
	// identifiers and literals
	IDENT = "IDENT"
	INT = "INT"
	FLOAT = "FLOAT"
	STRING = "STRING"
	BYTE = "BYTE"

	// operators
	ASSIGN = "="
	PLUS = "+"
	MINUS = "-"
	BANG = "!"
	ASTERISK = "*"
	SLASH = "/"
	LT = "<"
	GT = ">"
	EQ = "=="
	NOT_EQ = "!="
	MODULO = "%"

	// delimiters
	COMMA = ","
	SCOLON = ";"
	COLON = ":"
	OPAREN = "("
	CPAREN = ")"
	OBRACE = "{"
	CBRACE = "}"
	OBRACKET = "["
	CBRACKET = "]"

	// keywords
	FUNCTION = "FUNCTION"
	LET = "LET"
	TRUE = "TRUE"
	FALSE = "FALSE"
	AND = "AND"
	OR = "OR"
	IF = "IF"
	ELSE = "ELSE"
	RETURN = "RETURN"
)

var keywords = map[string]TokenType{
	"func": FUNCTION,
	"let": LET,
	"true": TRUE,
	"false": FALSE,
	"and": AND,
	"or": OR,
	"if": IF,
	"else": ELSE,
	"return": RETURN,
}

// if the identifier is a keyword, returns the keyword token
// else returns the identifier as the token
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
