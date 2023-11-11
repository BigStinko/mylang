package evaluator

import (
	"bytes"
	"fmt"
	"mylang/object"
	"os/exec"
	"regexp"
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

			switch args[0].Type() {
			case object.ARRAY_OBJ:
				array := args[0].(*object.Array)
				length := len(array.Elements)

				newElements := make([]object.Object, length + 1, length + 1)
				copy(newElements, array.Elements)
				newElements[length] = args[1]

				return &object.Array{Elements: newElements}
			case object.STRING_OBJ:
				if args[1].Type() != object.RUNE_OBJ {
					return newError("argument 2 to `push` must be RUNE, got %s",
						args[1].Type())
				}

				str := args[0].(*object.String)
				char := args[1].(*object.Rune)
				newStr := []rune(str.Value)
				newStr = append(newStr, char.Value)
				return &object.String{Value: string(newStr)}
			default:
				return newError("argument to `push` must be ARRAY or STRING, got %s",
					args[0].Type())

			}
		},	
	},

	"pop": {
		Function: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			switch args[0].Type() {
			case object.ARRAY_OBJ:
				array := args[0].(*object.Array)
				length := int64(len(array.Elements))
				
				if length == 0 {
					return NULL
				}

				out := array.Elements[length - 1]

				array.Elements = array.Elements[:length - 1]
				return out
			case object.STRING_OBJ:
				str := args[0].(*object.String)
				length := int64(len(str.Value))
				if length == 0 {
					return NULL
				}

				var b rune = []rune(str.Value)[length - 1]
				str.Value = str.Value[:length - 1]

				return &object.Rune{Value: b}
			default:
				return newError("argument to `assign` must be ARRAY or STRING, got %s",
					args[0].Type())
			}
		},
	},

	"assign": {
		Function: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3",
					len(args))
			}
			switch args[0].Type() {
			case object.ARRAY_OBJ:
				if args[1].Type() != object.INTEGER_OBJ {
					return newError("argument 2 to `assign` must be INTEGER, got %s",
						args[1].Type())
				}
				
				array := args[0].(*object.Array)
				length := int64(len(array.Elements))

				index := args[1].(*object.Integer).Value

				if (index > length - 1) {
					return newError("invalid index on array")
				}

				array.Elements[index] = args[2]
			
			case object.HASH_OBJ:
				hashKey, ok := args[1].(object.Hashable)
				if !ok {
					return newError("unusable as hash key: %s", args[1].Type())
				}

				hash := args[0].(*object.Hash)
				hash.Pairs[hashKey.HashKey()] = object.HashPair{Key: args[1], Value: args[2]}

			case object.STRING_OBJ:
				if args[1].Type() != object.INTEGER_OBJ {
					return newError("argument 2 to `assign` must be INTEGER, got %s",
						args[1].Type())
				}
				if args[2].Type() != object.RUNE_OBJ {
					return newError("argument 3 to `assign` must be BYTE, got %s",
						args[2].Type())
				}

				str := args[0].(*object.String)
				length := int64(len(str.Value))
				index := args[1].(*object.Integer).Value
				b := args[2].(*object.Rune)

				if (index > length - 1) {
					return newError("invalid index on string")
				}

				str.Value = str.Value[:index] + string(b.Value) + str.Value[index + 1 :] 
			default:
				return newError("argument 1 to `assign` must be ARRAY, HASH, or STRING got %s",
					args[0].Type())
			}

			return NULL
		},
	},

	"string": {
		Function: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}

			return &object.String{Value: args[0].Inspect()}
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

	"type": {
		Function: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			
			return &object.String{Value: string(args[0].Type())}
		},
	},

	"command": {
		Function: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			str, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `command` must be STRING. got=%q",
					args[0].Type())
			}

			fmt.Print("run")
			r := regexp.MustCompile(`[^\s"']+|"([^"]*)"|'([^']*)`)
			res := r.FindAllString(str.Value, -1)

			var result []string
			for _, e := range res {
				result = append(result, trimQuotes(e, '"'))
			}
			cmd := exec.Command(result[0], result[1:]...)

			var outb, errb bytes.Buffer
			cmd.Stdout = &outb
			cmd.Stderr = &errb
			err := cmd.Run()
			
			if err != nil && err != err.(*exec.ExitError) {
				fmt.Printf("%s' failed : %s\n", str.Value, err.Error())
				return NULL
			}

			stdout := &object.String{Value: outb.String()}
			stderr := &object.String{Value: errb.String()}

			stdoutKey := &object.String{Value: "stdout"}
			stdoutPair := object.HashPair{Key: stdoutKey, Value: stdout}

			stderrKey := &object.String{Value: "stderr"}
			stderrPair := object.HashPair{Key: stderrKey, Value: stderr}

			newHash := make(map[object.HashKey]object.HashPair)
			newHash[stdoutKey.HashKey()] = stdoutPair
			newHash[stderrKey.HashKey()] = stderrPair

			return &object.Hash{Pairs: newHash}
		},
	},
}

func trimQuotes(in string, c byte) string {
	if len(in) >= 2 {
		if in[0] == c && in[len(in) - 1] == c {
			return in[1 : len(in) - 1]
		}
	}
	return in
}
