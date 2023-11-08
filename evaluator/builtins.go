package evaluator

import (
	"fmt"
	"mylang/object"
)

var builtins = map[string]*object.Builtin{
	"len": {
		Function: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}

			switch arg := args[0].(type) {
			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			default:
				return newError("argument to `len` not supported, got %s",
					args[0].Type())
			}
		},
	},

	"last": {
		Function: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `last` must be ARRAY, got %s",
					args[0].Type())
			}

			array := args[0].(*object.Array)
			length := len(array.Elements)
			if length > 0 {
				return array.Elements[length - 1]
			}

			return NULL
		},
	},

	"rest": {
		Function: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `rest` must be ARRAY, got %s",
					args[0].Type())
			}
			
			array := args[0].(*object.Array)
			length := len(array.Elements)
			if length > 0 {
				newElements := make([]object.Object, length - 1, length - 1)
				copy(newElements, array.Elements[1:length])
				return &object.Array{Elements: newElements}
			}

			return NULL
		},
	},

	"push": {
		Function: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2",
					len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `push` must be ARRAY, got %s",
					args[0].Type())
			}
			
			array := args[0].(*object.Array)
			length := len(array.Elements)

			newElements := make([]object.Object, length + 1, length + 1)
			copy(newElements, array.Elements)
			newElements[length] = args[1]

			return &object.Array{Elements: newElements}
		},	
	},

	"puts": {
		Function: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Print(arg.Inspect())
			}

			return NULL
		},
	},
}
