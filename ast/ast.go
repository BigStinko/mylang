package ast

import (
	"mylang/token"
	"bytes"
)

type Node interface {
	TokenLiteral() string   // used only for testing/debugging
	String() string
}

type Statement interface {
	Node
	statementNode()   // dummy method to help distinguish statements from other nodes
}

type Expression interface {
	Node
	expressionNode()   // dummy method
}

type Program struct {
	Statements []Statement
}
// implements the Node interface
func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, statement := range p.Statements {
		out.WriteString(statement.String())
	}

	return out.String()
}


// identifiers have the identifier token an a string with their name
type Identifier struct {
	Token token.Token
	Value string
}
// implements the Expression and Node interface
func (i *Identifier) expressionNode() {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string { return i.Value }

// let statements have a let token, an identifier, and an expression
type LetStatement struct {
	Token token.Token
	Name *Identifier
	Value Expression
}
// implements the Statement and Node interface
func (ls *LetStatement) statementNode() {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }

func (ls *LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral() + " " )
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}

	out.WriteString(";")

	return out.String()
}

// return statements have the return token and a return value expression
type ReturnStatement struct {
	Token token.Token
	ReturnValue Expression
}
// implements the Statement and Node interface
func (rs *ReturnStatement) statementNode() {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }

func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")

	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}

	out.WriteString(";")

	return out.String()
}

// Token for the first token in the expression and the expression value
type ExpressionStatement struct {
	Token token.Token 
	Expression Expression
}
// implements the Statement and Node interface
func (es *ExpressionStatement) statementNode() {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

