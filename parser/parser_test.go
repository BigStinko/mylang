package parser

import (
	"testing"
	"mylang/ast"
	"mylang/lexer"
)

func checkParserErrors(t *testing.T, p *Parser) {
	var errors []string = p.Errors()
	
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))

	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

func testLetStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral not 'let'. got=%q", s.TokenLiteral())
	}

	letStatement, ok := s.(*ast.LetStatement)  // checks if the statement is LetStatement
	if !ok {
		t.Errorf("statement not *ast.LetStatement. got=%T", s)
		return false
	}

	if letStatement.Name.Value != name {
		t.Errorf("letstatement.Name.Value not '%s'. got=%s",
		name, letStatement.Name.Value)
		return false
	}

	if letStatement.Name.TokenLiteral() != name {
		t.Errorf("letstatement.Name.TokenLiteral() not '%s'. got=%s",
			name, letStatement.Name.TokenLiteral())
		return false
	}
	
	return true
}

func TestLetStatement(t *testing.T) {
	var input string = `
	let x = 5;
	let y = 10;
	let foobar = 838383;`

	l := lexer.New(input)
	var p *Parser = New(l)
	var program *ast.Program = p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got = %d", len(program.Statements))
	}
	
	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"foobar"},
	}

	for i, ts := range tests {
		var statement ast.Statement = program.Statements[i]
		if !testLetStatement(t, statement, ts.expectedIdentifier) {
			return
		}
	}
}

func TestReturnStatement(t *testing.T) {
	var input string = `
	return 5;
	return 10;
	return 993322;
	`

	l := lexer.New(input)
	var p *Parser = New(l)
	var program *ast.Program = p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d",
			len(program.Statements))
	}

	for _, statement := range program.Statements {
		returnStatement, ok := statement.(*ast.ReturnStatement) // checks if the statement is a ReturnStatement
		if !ok {
			t.Errorf("statement not *ast.ReturnStatement. got=%T",
			statement)
		}

		if returnStatement.TokenLiteral() != "return" {
			t.Errorf("returnStatement.TokenLiteral() not 'return', got=%q",
			returnStatement.TokenLiteral())
		}
	}
}
