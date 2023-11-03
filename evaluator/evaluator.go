package evaluator

import (
	"mylang/ast"
	"mylang/object"
	"mylang/token"
)

var (
	NULL = &object.Null{}
	TRUE = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func evaluateProgram(program *ast.Program) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = Evaluate(statement)

		if returnValue, ok := result.(*object.ReturnValue); ok {
			return returnValue.Value
		}
	}

	return result
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func getBoolObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func evaluateBangOperator(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evaluateMinusPrefixOperator(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return NULL
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evaluateIntegerInfixOperator(
	operator string,
	left, right object.Object,
) object.Object {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	switch operator {
	case token.PLUS:
		return &object.Integer{Value: leftValue + rightValue}
	case token.MINUS:
		return &object.Integer{Value: leftValue - rightValue}
	case token.ASTERISK:
		return &object.Integer{Value: leftValue * rightValue}
	case token.SLASH:
		return &object.Integer{Value: leftValue / rightValue}
	case token.LT:
		return getBoolObject(leftValue < rightValue)
	case token.GT:
		return getBoolObject(leftValue > rightValue)
	case token.EQ:
		return getBoolObject(leftValue == rightValue)
	case token.NOT_EQ:
		return getBoolObject(leftValue != rightValue)
	default:
		return NULL
	}
}

func evaluatePrefixOperator(operator string, right object.Object) object.Object {
	switch operator {
	case token.BANG:
		return evaluateBangOperator(right)
	case token.MINUS:
		return evaluateMinusPrefixOperator(right)
	default:
		return NULL
	}
}

func evaluateInfixOperator(
	operator string,
	left, right object.Object,
) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evaluateIntegerInfixOperator(operator, left, right)
	case operator == token.EQ:
		return getBoolObject(left == right)
	case operator == token.NOT_EQ:
		return getBoolObject(left != right)
	default:
		return NULL
	}
}

func evaluateIfExpression(ie *ast.IfExpression) object.Object {
	if isTruthy(Evaluate(ie.Condition)) {
		return Evaluate(ie.Consequence)
	} else if ie.Alternative != nil {
		return Evaluate(ie.Alternative)
	} else {
		return NULL
	}
}

func evaluateBlockStatements(block *ast.BlockStatement) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Evaluate(statement)
		
		if result != nil && result.Type() == object.RETURN_VALUE_OBJ {
			return result
		}
	}

	return result
}

func Evaluate(node ast.Node) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evaluateProgram(node)
	case *ast.BlockStatement:
		return evaluateBlockStatements(node)
	case *ast.ExpressionStatement:
		return Evaluate(node.Expression)
	case *ast.PrefixExpression:
		return evaluatePrefixOperator(node.Operator, Evaluate(node.Right))
	case *ast.InfixExpression:
		left := Evaluate(node.Left)
		right := Evaluate(node.Right)
		return evaluateInfixOperator(node.Operator, left, right)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.BooleanLiteral:
		return getBoolObject(node.Value)
	case *ast.IfExpression:
		return evaluateIfExpression(node)
	}

	return nil
}
