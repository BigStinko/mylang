package parser

import (
	"fmt"
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

func TestIdentifierExpression(t *testing.T) {
	var input string = "foobar"

	l := lexer.New(input)
	var p *Parser = New(l)
	var program *ast.Program = p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statements, got=%d",
			len(program.Statements))
	}

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)

	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			statement.Expression)
	}

	identifier, ok := statement.Expression.(*ast.Identifier)

	if !ok {
		t.Fatalf("expression not *ast.Identifier. got=%T", statement.Expression)
	}
	if identifier.Value != "foobar" {
		t.Errorf("identifier.Value not %s. got=%s", "foobar", identifier.Value)
	}
	if identifier.TokenLiteral() != "foobar" {
		t.Errorf("identifier.TokenLiteral() not %s. got=%s",
			"foobar", identifier.TokenLiteral())
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	var input string = "5;"

	l := lexer.New(input)
	var p *Parser = New(l)
	var program *ast.Program = p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program does not have enough statements. got=%d",
			len(program.Statements))
	}

	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	literal, ok := statement.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expression not *ast.integerLiteral. got=%T", statement.Expression)
	}
	if literal.Value != 5 {
		t.Errorf("literal.Value not %d. got=%d", 5, literal.Value)
	}
	if literal.TokenLiteral() != "5" {
		t.Errorf("literal.TokenLiteral() not %s. got=%s", "5",
			literal.TokenLiteral())
	}
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integer, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("integer literal not *ast.IntegerLiteral. got=%T", il)
		return false
	}
	if integer.Value != value {
		t.Errorf("integer.Value not %d. got=%d", value, integer.Value)
		return false
	}
	if integer.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("integer.TokenLiteral not %d. got=%s",
			value, integer.TokenLiteral())
		return false
	}

	return true
}

func TestParsingPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input string
		operator string
		integerValue int64
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
	}

	for _, prefixTest := range prefixTests {
		l := lexer.New(prefixTest.input)
		var p *Parser = New(l)
		var program *ast.Program = p.ParseProgram()
		checkParserErrors(t, p)
		
		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
				1, len(program.Statements))
		}

		statement, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}

		expression, ok := statement.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("statment is not ast.PrefixExpression. got=%T", statement.Expression)
		}
		if expression.Operator != prefixTest.operator {
			t.Fatalf("expression.Operator is not '%s'. got=%s",
				prefixTest.operator, expression.Operator)
		}
		if !testIntegerLiteral(t, expression.Right, prefixTest.integerValue) {
			return
		}
	}
}

func TestParsingInfixExpression(t *testing.T) {
	infixTests := []struct {
		input string
		leftValue int64
		operator string
		rightValue int64
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
	}

	for _, infixTest := range infixTests {
		l := lexer.New(infixTest.input)
		var p *Parser = New(l)
		var program *ast.Program = p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
				1, len(program.Statements))
		}

		statement, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}

		expression, ok := statement.Expression.(*ast.InfixExpression)
		if !ok {
			t.Fatalf("expression is not ast.InfixExpression. Got=%T",
				statement.Expression)
		}

		if !testIntegerLiteral(t, expression.Left, infixTest.leftValue) {
			return
		}

		if expression.Operator != infixTest.operator {
			t.Fatalf("expression.Operator is not '%s', got=%s",
				infixTest.operator, expression.Operator)
		}

		if !testIntegerLiteral(t, expression.Right, infixTest.rightValue) {
			return
		}
	}

}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input string
		exprected string
	}{
		{
			"-a * b",
			"((-a) * b)",
		},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},
		{
			"a + b - c",
			"((a + b) - c)",
		},
		{
			"a * b * c",
			"((a * b) * c)",
		},
		{
			"a * b / c",
			"((a * b) / c)",
		},
		{
			"a + b / c",
			"(a + (b / c))",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f)",
		},
		{
			"3 + 4; -5 * 5",
			"(3 + 4)((-5) * 5)",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4))",
		},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
	}

	for _, test := range tests {
		l := lexer.New(test.input)
		var p *Parser = New(l)
		var program *ast.Program = p.ParseProgram()
		checkParserErrors(t, p)
		
		var actual string = program.String()
		if actual != test.exprected {
			t.Errorf("exprected=%q, got=%q", test.exprected, actual)
		}
	}
}
