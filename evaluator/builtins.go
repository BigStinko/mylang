package evaluator

import (
	"bufio"
	"bytes"
	"fmt"
	"mylang/object"
	"os"
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
				if len(e) >= 2 {
					if e[0] == '"' && e[len(e) - 1] == '"' {
						e = e[1 : len(e) - 1]
					}
				}
				result = append(result, e)
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

	"open": {
		Function: func(args ...object.Object) object.Object {
			if len(args) > 2 {
				return newError("wrong number of arguments. got=%d, want=2",
					len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument 1 to `open` must be STRING. got=%q",
					args[0].Type())
			}
			var mode string
			if len(args) > 1 {
				if args[1].Type() != object.STRING_OBJ {
					return newError("argument 2 to `open` must be STRING. got=%q",
						args[1].Type())
				}

				mode = args[1].(*object.String).Value
			}

			var path string = args[0].(*object.String).Value

			md := os.O_RDONLY

			if mode == "w" {
				md = os.O_WRONLY
				os.Remove(path)
			} else if mode == "wa" || mode == "aw" {
				md = os.O_WRONLY | os.O_APPEND
			}

			file, err := os.OpenFile(path, os.O_CREATE | md, 0644)
			if err != nil {
				return newError(err.Error())
			}

			fileObj := &object.File{Handle: file, Path: path}

			if md == os.O_RDONLY {
				fileObj.Reader = bufio.NewReader(file)
			} else {
				fileObj.Writer = bufio.NewWriter(file)
			}

			return fileObj
		},
	},

	"close": {
		Function: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != object.FILE_OBJ {
				return newError("argument to `close` must be FILE. got=%q",
					args[0].Type())
			}

			file := args[0].(*object.File).Handle
			file.Close()
			return TRUE
		},	
	},

	"read": {
		Function: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != object.FILE_OBJ {
				return newError("argument to `read` must be FILE. got=%q",
					args[0].Type())
			}

			file := args[0].(*object.File)
			if file.Reader == nil {
				return &object.String{Value: ""}
			}

			out, err := file.Reader.ReadString('\n')
			if err != nil {
				return &object.String{Value: out}
			}

			return &object.String{Value: out}
		},	
	},
	
	"write": {
		Function: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2",
					len(args))
			}
			if args[0].Type() != object.FILE_OBJ {
				return newError("argument 1 to `write` must be FILE. got=%q",
					args[0].Type())
			}

			file := args[0].(*object.File)

			if file.Writer == nil {
				return FALSE
			}

			_, err := file.Writer.Write([]byte(args[1].Inspect()))
			if err == nil {
				file.Writer.Flush()
				return TRUE
			}

			return FALSE
		},	
	},

	"remove": {
		Function: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != object.FILE_OBJ {
				return newError("argument 1 to `remove` must be FILE. got=%q",
					args[0].Type())
			}

			file := args[0].(*object.File)

			err := os.Remove(file.Path)
			if err != nil {
				return &object.String{Value: err.Error()}
			}

			return TRUE
		},
	},

	"args": {
		Function: func(args ...object.Object) object.Object {

			switch len(args) {
			case 0:
				length := len(os.Args[1:])
				out := make([]object.Object, length)
				for i, arg := range os.Args[1:] {
				out[i] = &object.String{Value: arg}
				}
				return &object.Array{Elements: out}
			case 1:
				if args[0].Type() != object.INTEGER_OBJ {
					return newError("argument to `args` must be INTEGER. got=%q",
						args[0].Type())
				}
				index := args[0].(*object.Integer).Value
				osArgs := os.Args[1:]
				if index > int64(len(osArgs)) - 1 || index < 0 {
					return newError("out of bounds index")
				}
				return &object.String{Value: osArgs[index]}
			default:
				return newError("wrong number of arguments. got=%d, want 0 or 1",
					len(args))
			}
		},
	},
}
