package parser

import (
	"fmt"
	"mylang/ast"
	"mylang/lexer"
	"mylang/token"
)

type Parser struct {
	l *lexer.Lexer
	currentToken token.Token
	nextToken token.Token
	errors []string
}

func (p *Parser) advanceTokens() {
	p.currentToken = p.nextToken
	p.nextToken = p.l.NextToken()
}

func New(l *lexer.Lexer) *Parser {
	var p *Parser = &Parser{
		l: l,
		errors: []string{},
	}

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

func (p *Parser) expectedToken(t token.TokenType) bool {
	if p.nextToken.Type == t {
		p.advanceTokens()
		return true
	} else {
		p.expectedTokenError(t)
		return false
	}
}

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

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	statement := &ast.ReturnStatement{Token: p.currentToken}

	p.advanceTokens()

	for p.currentToken.Type != token.SCOLON {
		p.advanceTokens()
	}

	return statement
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.currentToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return nil
	}
}

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
