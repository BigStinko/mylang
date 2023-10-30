package parser

import (
	"fmt"
	"mylang/ast"
	"mylang/lexer"
	"mylang/token"
	"strconv"
)

type Parser struct {
	l *lexer.Lexer
	errors []string  // for holding multiple erros instead of halting on every error
	// current and next token for identifying different kinds of expressions and statements
	currentToken token.Token
	nextToken token.Token
	// map of prefix and infix functions that hold the specific function for every expression
	prefixParseFunctions map[token.TokenType]prefixParseFunction
	infixParseFunctions map[token.TokenType]infixParseFunction
}

// types of fucntions for the parser with prefixParseFunction being
// for urnary operators, literals, and identifiers 
// infixParseFunction being for binary operators
type (
	prefixParseFunction func() ast.Expression
	infixParseFunction func(ast.Expression) ast.Expression
)

// orders the precedence of operations for resolving prefix and infix expressions
const (
	_ int = iota
	LOWEST
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
	CALL
)

var precedences = map[token.TokenType]int{
	token.EQ: EQUALS,
	token.NOT_EQ: EQUALS,
	token.LT: LESSGREATER,
	token.GT: LESSGREATER,
	token.PLUS: SUM,
	token.MINUS: SUM,
	token.SLASH: PRODUCT,
	token.ASTERISK: PRODUCT,
}

// used by the parser to walk through the sequence of tokens.
// also advances the tokens in the lexer
func (p *Parser) advanceTokens() {
	p.currentToken = p.nextToken
	p.nextToken = p.l.NextToken()
}

// prefix parse function for identifier tokens. Makes and returns an identifier expression
// from the current token in the parser
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}
}

// prefix parse function for integer literals. Makes a integer literal expression, and
// gets the value of the integer from the current token, and returns the expression
func (p *Parser) parseIntegerLiteral() ast.Expression {
	literal := &ast.IntegerLiteral{Token: p.currentToken}

	value, err := strconv.ParseInt(p.currentToken.Literal, 0, 64)
	if err != nil {
		var msg string = fmt.Sprintf("could not parse %q as integer", p.currentToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	literal.Value = value

	return literal
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token: p.currentToken,
		Operator: p.currentToken.Literal,
	}

	p.advanceTokens()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) nextPrecedence() int {
	if precedence, ok := precedences[p.nextToken.Type]; ok {
		return precedence
	}
	return LOWEST
}

func (p *Parser) currentPrecedence() int {
	if precedence, ok := precedences[p.currentToken.Type]; ok {
		return precedence
	}

	return LOWEST
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token: p.currentToken,
		Operator: p.currentToken.Literal,
		Left: left,
	}
	
	var precedence int = p.currentPrecedence()
	p.advanceTokens()
	expression.Right = p.parseExpression(precedence)

	return expression
}

// makes and returns a parser for the given lexer
func New(l *lexer.Lexer) *Parser {
	var p *Parser = &Parser{
		l: l,
		errors: []string{},
	}

	p.prefixParseFunctions = make(map[token.TokenType]prefixParseFunction)
	p.prefixParseFunctions[token.IDENT] = p.parseIdentifier
	p.prefixParseFunctions[token.INT] = p.parseIntegerLiteral
	p.prefixParseFunctions[token.BANG] = p.parsePrefixExpression
	p.prefixParseFunctions[token.MINUS] = p.parsePrefixExpression

	p.infixParseFunctions = make(map[token.TokenType]infixParseFunction)
	p.infixParseFunctions[token.PLUS] = p.parseInfixExpression
	p.infixParseFunctions[token.MINUS] = p.parseInfixExpression
	p.infixParseFunctions[token.SLASH] = p.parseInfixExpression
	p.infixParseFunctions[token.ASTERISK] = p.parseInfixExpression
	p.infixParseFunctions[token.EQ] = p.parseInfixExpression
	p.infixParseFunctions[token.NOT_EQ] = p.parseInfixExpression
	p.infixParseFunctions[token.LT] = p.parseInfixExpression
	p.infixParseFunctions[token.GT] = p.parseInfixExpression

	p.advanceTokens()
	p.advanceTokens()

	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) expectedTokenError(t token.TokenType) {
	var msg string = fmt.Sprintf("exprected next token to be %s, got %s instead",
		t, p.nextToken.Type)
	p.errors = append(p.errors, msg)
}

// used by the parser to evaluate the order of tokens in a statement/expression
// determines if the expected token is the next token in the program
func (p *Parser) expectedToken(t token.TokenType) bool {
	if p.nextToken.Type == t {
		p.advanceTokens()
		return true
	} else {
		p.expectedTokenError(t)
		return false
	}
}

// used by the parseStatement method to parse a let statement in the program
// generates a LetStatement Node and gets the identifier and expression fields
// from the program and ends at the semicolon
func (p *Parser) parseLetStatement() *ast.LetStatement {
	statement := &ast.LetStatement{Token: p.currentToken}

	if !p.expectedToken(token.IDENT) {
		return nil
	}

	statement.Name = &ast.Identifier{
		Token: p.currentToken,
		Value: p.currentToken.Literal,
	}

	if !p.expectedToken(token.ASSIGN) {
		return nil
	}

	for p.currentToken.Type != token.SCOLON {
		p.advanceTokens()
	}
	
	return statement
}

// used by the parseStatement method to parse return statements in the program
// generatees a ReturnStatement Node and attaches the return expression
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	statement := &ast.ReturnStatement{Token: p.currentToken}

	p.advanceTokens()

	for p.currentToken.Type != token.SCOLON {
		p.advanceTokens()
	}

	return statement
}

func (p *Parser) noPRefixParseFunctionError(t token.TokenType) {
	var msg string = fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFunctions[p.currentToken.Type]

	if prefix == nil {
		p.noPRefixParseFunctionError(p.currentToken.Type)
		return nil
	}

	leftExpression := prefix()

	for p.nextToken.Type != token.SCOLON && precedence < p.nextPrecedence() {
		infix := p.infixParseFunctions[p.nextToken.Type]

		if infix == nil {
			return leftExpression
		}

		p.advanceTokens()

		leftExpression = infix(leftExpression)
	}

	return leftExpression
}

// generates the ExpressionStatement Node for the following statment in the
// program
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	statement := &ast.ExpressionStatement{Token: p.currentToken}

	statement.Expression = p.parseExpression(LOWEST)

	if p.nextToken.Type == token.SCOLON {
		p.advanceTokens()
	}

	return statement
}

// used by the ParseProgram method to parse the next statement in the program
// parses let, return, and expression statments
func (p *Parser) parseStatement() ast.Statement {
	switch p.currentToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// generates the AST for the program by generating statement nodes for
// every statment in the input.
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.currentToken.Type != token.EOF {
		var statement ast.Statement = p.parseStatement()
		if statement != nil {
			program.Statements = append(program.Statements, statement)
		}
		p.advanceTokens()
	}

	return program
}
