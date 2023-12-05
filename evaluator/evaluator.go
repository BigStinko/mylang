package evaluator

import (
	"fmt"
	"math"
	"mylang/ast"
	"mylang/object"
	"mylang/token"
)

var builtins = map[string]*object.Builtin{
	"len": object.GetBuiltinByName("len"),
	"puts": object.GetBuiltinByName("puts"),
	"first": object.GetBuiltinByName("first"),
	"last": object.GetBuiltinByName("last"),
	"rest": object.GetBuiltinByName("rest"),
	"push": object.GetBuiltinByName("push"),
	"pop": object.GetBuiltinByName("pop"),
	"string": object.GetBuiltinByName("string"),
	"keys": object.GetBuiltinByName("keys"),
	"delete": object.GetBuiltinByName("delete"),
	"assign": object.GetBuiltinByName("assign"),
	"type": object.GetBuiltinByName("type"),
	"command": object.GetBuiltinByName("command"),
	"open": object.GetBuiltinByName("open"),
	"close": object.GetBuiltinByName("close"),
	"read": object.GetBuiltinByName("read"),
	"write": object.GetBuiltinByName("write"),
	"remove": object.GetBuiltinByName("remove"),
	"args": object.GetBuiltinByName("args"),
}

// recursively evaluates every kind of node in the ast
func Evaluate(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {

	// statements
	case *ast.Program:
		return evaluateProgram(node, env)

	case *ast.BlockStatement:
		return evaluateBlockStatement(node, env)
	
	case *ast.ReturnStatement:
		value := Evaluate(node.ReturnValue, env)
		if isError(value) {
			return value
		}
		return &object.ReturnValue{Value: value}
	
	case *ast.LetStatement:
		// connects a value with an identifier in the environment
		value := Evaluate(node.Value, env)
		if isError(value) {
			return value
		}
		env.Set(node.Name.Value, value)
	
	case *ast.AssignmentStatement:
		value := Evaluate(node.Value, env)
		if isError(value) {
			return value
		}

		if !env.Assign(node.Name.Value, value) {
			return newError("undefined identifier '%s'", node.Name.Value)
		}

		return value
	
	case *ast.ExpressionStatement:
		return Evaluate(node.Expression, env)
	
	// expressions
	case *ast.PrefixExpression:
		right := Evaluate(node.Right, env)
		if isError(right) {
			return right
		}
		return evaluatePrefixOperator(node.Token, right)
	
	case *ast.InfixExpression:
		left := Evaluate(node.Left, env)
		if isError(left) {
			return left
		}
		right := Evaluate(node.Right, env)
		if isError(right) {
			return right
		}
		return evaluateInfixOperator(node.Token, left, right)
	
	case *ast.Identifier:
		return evaluateIdentifier(node, env)
	
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	
	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}

	case *ast.StringLiteral:
		return &object.String{Value: node.Value}

	case *ast.BooleanLiteral:
		return getBoolObject(node.Value)
	
	case *ast.ArrayLiteral:
		elements := evaluateExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}

	case *ast.HashLiteral:
		return evaluateHashLiteral(node, env)

	case *ast.IfExpression:
		return evaluateIfExpression(node, env)
	
	case *ast.WhileExpression:
		return evaluateWhileExpression(node, env)
	
	case *ast.SwitchExpression:
		return evaluateSwitchExpression(node, env)

	case *ast.FunctionLiteral:
		// creates a function object with the function literal node's 
		// parameters and body plus the outer environment for the node
		return &object.Function{
			Parameters: node.Parameters,
			Environment: env,
			Body: node.Body,
		}
	
	case *ast.CallExpression:
		// evaluates a function call expression 
		function := Evaluate(node.Function, env)
		if isError(function) {
			return function
		}
		args := evaluateExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args)
	
	case *ast.IndexExpression:
		left := Evaluate(node.Left, env)
		if isError(left) {
			return left
		}
		index := Evaluate(node.Index, env)
		if isError(index) {
			return index
		}
		return evaluateIndexExpression(left, index)
	}

	return nil
}

// calls evaluate on every statment in a program. If it encounters
// a returnValue then it stops evaluation there. Used by Evaluate 
// function to evaluate a program node
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

// similar to evaluateProgram, but instead of returning the value of
// return value objects keeps them as the return value object so
// return values in nested block can still cause an outer statement 
// to return
func evaluateBlockStatement(
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

// used to check if an object is an error so functions know when to halt
// their execution and return the error
func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

// returns the opposite truth value of the given object. Opposite of
// isTruthy. Presently returns the opposite of booleans and if an object
// is not NULL then returns true
func evaluateBangOperator(right object.Object) object.Object {
	switch right {
	case object.TRUE:
		return object.FALSE
	case object.FALSE:
		return object.TRUE
	case object.NULL:
		return object.TRUE
	default:
		return object.FALSE
	}
}

// determines the truth value of the given object. As long as the
// object has a value that is not NULL or FALSE, it returns true
func isTruthy(obj object.Object) bool {
	switch obj {
	case object.NULL:
		return false
	case object.TRUE:
		return true
	case object.FALSE:
		return false
	default:
		return true
	}
}

// returns a new error object. Uses the same interface as Sprintf
func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

// constructs and returns an integer object with the opposite value
func evaluateMinusPrefixOperator(right object.Object) object.Object {
	switch right.Type() {
	case object.INTEGER_OBJ:
		value := right.(*object.Integer).Value
		return &object.Integer{Value: -value}
	case object.FLOAT_OBJ:
		value := right.(*object.Float).Value
		return &object.Float{Value: -value}
	default:
		return newError("unknown operator: -%s", right.Type())
	}
}

// evaluates the two prefix operators
func evaluatePrefixOperator(
	operator token.Token,
	right object.Object,
) object.Object {
	switch operator.Type {
	case token.BANG:
		return evaluateBangOperator(right)
	case token.MINUS:
		return evaluateMinusPrefixOperator(right)
	default:
		return newError("unknown operator: %s%s",
			operator.Literal, right.Type())
	}
}

// returns the universal bool objects from the input boolean
func getBoolObject(input bool) *object.Boolean {
	if input {
		return object.TRUE
	}
	return object.FALSE
}

// evaluates all the binary operators for integers.
func evaluateIntegerInfixOperator(
	operator token.Token,
	left, right object.Object,
) object.Object {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	switch operator.Type {
	case token.PLUS:
		return &object.Integer{Value: leftValue + rightValue}
	case token.MINUS:
		return &object.Integer{Value: leftValue - rightValue}
	case token.ASTERISK:
		return &object.Integer{Value: leftValue * rightValue}
	case token.SLASH:
		return &object.Integer{Value: leftValue / rightValue}
	case token.MODULO:
		return &object.Integer{Value: leftValue % rightValue}
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
			left.Type(), operator.Literal, right.Type())
	}
}

// evaluates all the binary operators for floats.
func evaluateFloatInfixOperator(
	operator token.Token,
	left, right object.Object,
) object.Object {
	leftValue := left.(*object.Float).Value
	rightValue := right.(*object.Float).Value

	switch operator.Type {
	case token.PLUS:
		return &object.Float{Value: leftValue + rightValue}
	case token.MINUS:
		return &object.Float{Value: leftValue - rightValue}
	case token.ASTERISK:
		return &object.Float{Value: leftValue * rightValue}
	case token.SLASH:
		return &object.Float{Value: leftValue / rightValue}
	case token.MODULO:
		return &object.Float{Value: math.Mod(leftValue, rightValue)}
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
			left.Type(), operator.Literal, right.Type())
	}
}

func evaluateStringInfixOperator(
	operator token.Token,
	left, right object.Object,
) object.Object {
	if operator.Type != token.PLUS {
		return newError("unknown operator: %s %s %s",
			left.Type(), operator.Literal, right.Type())
	}

	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value

	return &object.String{Value: leftValue + rightValue}
}

// evaluates all general binary operators
func evaluateInfixOperator(
	operator token.Token,
	left, right object.Object,
) object.Object {
	switch {
	case operator.Type == token.AND:
		return getBoolObject(isTruthy(left) && isTruthy(right))
	
	case operator.Type == token.OR:
		return getBoolObject(isTruthy(left) || isTruthy(right))
	
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evaluateIntegerInfixOperator(operator, left, right)
	
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.FLOAT_OBJ:
		return evaluateFloatInfixOperator(operator, left, right)
	
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evaluateStringInfixOperator(operator, left, right)
	
	case operator.Type == token.EQ:
		return getBoolObject(left == right)
	
	case operator.Type == token.NOT_EQ:
		return getBoolObject(left != right)
	
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s",
			left.Type(), operator.Literal, right.Type())
	
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator.Literal, right.Type())
	}
}

// checks if the identifier is in the enviroment, and returns the value if so
func evaluateIdentifier(
	node *ast.Identifier,
	env *object.Environment,
) object.Object {
	if value, ok := env.Get(node.Value); ok {
		return value
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return newError("identifier not found: " + node.Value)
}

// evaluates if and if else expressions. If the condition is true
// evaluates the consequence statement, and if the alternative
// statement exists evaluates it
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
		return object.NULL
	}
}

func evaluateWhileExpression(
	we *ast.WhileExpression,
	env *object.Environment,
) object.Object {
	returnValue := &object.Boolean{Value: true}

	for {
		condition := Evaluate(we.Condition, env)
		if isError(condition) {
			return condition
		}
		if isTruthy(condition) {
			returnValue := Evaluate(we.Body, env)
			if returnValue.Type() == object.RETURN_VALUE_OBJ || returnValue.Type() == object.ERROR_OBJ {
				return returnValue
			}
		} else {
			break
		}
	}
	return returnValue
}

func evaluateSwitchExpression(
	se *ast.SwitchExpression,
	env *object.Environment,
) object.Object {
	value := Evaluate(se.Value, env)
	if isError(value) {
		return value
	}

	for _, choice := range se.Cases {
		if choice.Default {
			continue
		}

		out := Evaluate(choice.Value, env)
		if isError(out) {
			return out
		}

		if value.Type() == out.Type() && value.Inspect() == out.Inspect() {
			return evaluateBlockStatement(choice.Body, env)
		}
	}

	for _, choice := range se.Cases {
		if choice.Default {
			return evaluateBlockStatement(choice.Body, env)
		}
	}

	return nil
}

// evaluates a slice of expressions and returns a corresponding slice of
// the resulting objects from the given enviroment. Used to evaluate the
// parameters of a function call
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

// creates an extended environment for the function and adds the function 
// parameters to the environment. Now when the function executes it will
// have access to the values in the scope it was created in, as well as 
// any values created in the scope of the function. However those values
// will be restricted to the enclosed environment
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

// executes the given function object with the given arguments. Extends the 
// environment with the given arguments, evaluates the function, and returns
// the return value
func applyFunction(
	fn object.Object,
	args []object.Object,
) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnvironment(fn, args)
		evaluated := Evaluate(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	
	case *object.Builtin:
		if result := fn.Function(args...); result != nil {
			return result
		}
		return object.NULL
	default:
		return newError("not a function: %s", fn.Type())
	}
}

func evaluateArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	maximum := int64(len(arrayObject.Elements) - 1)

	if idx < 0 || idx > maximum {
		return object.NULL
	}

	return arrayObject.Elements[idx]
}

func evaluateStringIndexExpression(str, index object.Object) object.Object {
	strObj := str.(*object.String)
	idx := index.(*object.Integer).Value
	maximum := int64(len(strObj.Value) - 1)

	if idx < 0 || idx > maximum {
		return object.NULL
	}

	return &object.String{Value: string(strObj.Value[idx])}
}

func evaluateHashIndexExpression(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}
	
	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return object.NULL
	}

	return pair.Value
}

func evaluateIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evaluateArrayIndexExpression(left, index)
	case left.Type() == object.STRING_OBJ && index.Type() == object.INTEGER_OBJ:
		return evaluateStringIndexExpression(left, index)
	case left.Type() == object.HASH_OBJ:
		return evaluateHashIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evaluateHashLiteral(
	node *ast.HashLiteral,
	env *object.Environment,
) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := Evaluate(keyNode, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}

		value := Evaluate(valueNode, env)
		if isError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}

	return &object.Hash{Pairs: pairs}
}
