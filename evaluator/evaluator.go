package evaluator

import (
	"fmt"
	"mylang/ast"
	"mylang/object"
	"mylang/token"
)

var (
	NULL = &object.Null{}
	TRUE = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

// calls evaluate on every statment in a program. If it encounters
// a returnValue then it stops evaluation there
func evaluateProgram(
	program *ast.Program,
	env *object.Environment,
) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = Evaluate(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

// similare to evaluateProgram, but instead of returning the value of
// return value objects keeps them as the return value object so 
func evaluateBlockStatements(
	block *ast.BlockStatement,
	env *object.Environment,
) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Evaluate(statement, env)
		
		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func evaluateIdentifier(
	node *ast.Identifier,
	env *object.Environment,
) object.Object {
	value, ok := env.Get(node.Value)
	if !ok {
		return newError("identifier not found: " + node.Value)
	}
	return value
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

// determines the truth value of the given object. As long as the
// object has a value that is not NULL or FALSE, it returns true
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

// returns the universal bool objects from the input boolean
func getBoolObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

// returns the opposite truth value of the given object. Opposite of
// isTruthy
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

// constructs and returns an integer object with the opposite value
func evaluateMinusPrefixOperator(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

// evaluates all the binary operators for integers.
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
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

// evaluates the two prefix operators
func evaluatePrefixOperator(operator string, right object.Object) object.Object {
	switch operator {
	case token.BANG:
		return evaluateBangOperator(right)
	case token.MINUS:
		return evaluateMinusPrefixOperator(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

// evaluates all binary operators
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
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

// evaluates if and if else expressions. If the condition is true
// evaluates the consequence statement, and if the alternative
// statement exists evaluates it otherwise
func evaluateIfExpression(
	ie *ast.IfExpression,
	env *object.Environment,
) object.Object {
	condition := Evaluate(ie.Condition, env)
	if isError(condition) {
		return condition
	}
	if isTruthy(condition) {
		return Evaluate(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Evaluate(ie.Alternative, env)
	} else {
		return NULL
	}
}

func evaluateExpressions(
	expressions []ast.Expression,
	env *object.Environment,
) []object.Object {
	var result []object.Object

	for _, expression := range expressions {
		evaluated := Evaluate(expression, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}

	return obj
}

func extendFunctionEnvironment(
	fn *object.Function,
	args []object.Object,
) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Environment)

	for i, parameter := range fn.Parameters {
		env.Set(parameter.Value, args[i])
	}

	return env
}

func applyFunction(
	fn object.Object,
	args []object.Object,
) object.Object {
	function, ok := fn.(*object.Function)
	if !ok {
		return newError("not a function: %s", fn.Type())
	}

	extendedEnv := extendFunctionEnvironment(function, args)
	evaluated := Evaluate(function.Body, extendedEnv)
	return unwrapReturnValue(evaluated)
}

// recursively evaluates every kind of node in the ast
func Evaluate(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {

	// statements
	case *ast.Program:
		return evaluateProgram(node, env)

	case *ast.BlockStatement:
		return evaluateBlockStatements(node, env)
	
	case *ast.ReturnStatement:
		value := Evaluate(node.ReturnValue, env)
		if isError(value) {
			return value
		}
		return &object.ReturnValue{Value: value}
	
	case *ast.LetStatement:
		value := Evaluate(node.Value, env)
		if isError(value) {
			return value
		}
		env.Set(node.Name.Value, value)
	
	case *ast.ExpressionStatement:
		return Evaluate(node.Expression, env)
	
	// expressions
	case *ast.PrefixExpression:
		right := Evaluate(node.Right, env)
		if isError(right) {
			return right
		}
		return evaluatePrefixOperator(node.Operator, right)
	case *ast.InfixExpression:
		left := Evaluate(node.Left, env)
		if isError(left) {
			return left
		}
		right := Evaluate(node.Right, env)
		if isError(right) {
			return right
		}
		return evaluateInfixOperator(node.Operator, left, right)
	
	case *ast.Identifier:
		return evaluateIdentifier(node, env)
	
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	
	case *ast.BooleanLiteral:
		return getBoolObject(node.Value)
	
	case *ast.IfExpression:
		return evaluateIfExpression(node, env)
	
	case *ast.FunctionLiteral:
		return &object.Function{
			Parameters: node.Parameters,
			Environment: env,
			Body: node.Body,
		}
	
	case *ast.CallExpression:
		function := Evaluate(node.Function, env)
		if isError(function) {
			return function
		}
		args := evaluateExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args)
	}

	return nil
}
