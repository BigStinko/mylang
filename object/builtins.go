package object

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

var (
	TRUE = &Boolean{Value: true}
	FALSE = &Boolean{Value: false}
	NULL = &Null{}
)

var Builtins = []struct {
	Name    string
	Builtin *Builtin
}{
	{
		"len",
		&Builtin{Function: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments to `len`. got=%d, want=1",
					len(args))
			}

			switch arg := args[0].(type) {
			case *Array:
				return &Integer{Value: int64(len(arg.Elements))}
			case *String:
				return &Integer{Value: int64(len(arg.Value))}
			default:
				return newError("argument to `len` not supported, got %s",
					args[0].Type())
			}
		}},
	},
	{
		"puts",
		&Builtin{Function: func(args ...Object) Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}

			return NULL
		}},
	},
	{
		"first",
		&Builtin{Function: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments to `first`. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("argument to `first` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*Array)
			if len(arr.Elements) > 0 {
				return arr.Elements[0]
			}

			return NULL
		}},
	},
	{
		"last",
		&Builtin{Function: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments to `last`. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("argument to `last` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*Array)
			length := len(arr.Elements)
			if length > 0 {
				return arr.Elements[length-1]
			}

			return NULL
		}},
	},
	{
		"rest",
		&Builtin{Function: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments to `rest`. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("argument to `rest` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*Array)
			length := len(arr.Elements)
			if length > 0 {
				newElements := make([]Object, length-1, length-1)
				copy(newElements, arr.Elements[1:length])
				return &Array{Elements: newElements}
			}

			return NULL
		}},
	},
	{
		"push",
		&Builtin{Function: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments to `push`. got=%d, want=2",
					len(args))
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("argument to `push` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*Array)
			length := len(arr.Elements)

			newElements := make([]Object, length+1, length+1)
			copy(newElements, arr.Elements)
			newElements[length] = args[1]

			return &Array{Elements: newElements}
		}},
	},
	{
		"pop",
		&Builtin{Function: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments to `pop`. got=%d, want=1",
					len(args))
			}
			switch args[0].Type() {
			case ARRAY_OBJ:
				array := args[0].(*Array)
				length := int64(len(array.Elements))
				
				if length == 0 {
					return NULL
				}

				out := array.Elements[length - 1]

				array.Elements = array.Elements[:length - 1]
				return out
			case STRING_OBJ:
				str := args[0].(*String)
				length := int64(len(str.Value))
				if length == 0 {
					return NULL
				}

				s := str.Value[length - 1]
				str.Value = str.Value[:length - 1]

				return &String{Value: string(s)}
			default:
				return newError("argument to `assign` must be ARRAY or STRING, got %s",
					args[0].Type())
			}
		}},
	},
	{
		"string",
		&Builtin{Function: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments to `string`. got=%d, want=1",
					len(args))
			}

			return &String{Value: args[0].Inspect()}
		}},
	},
	{
		"keys", 
		&Builtin{Function: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments to `keys`. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != HASH_OBJ {
				return newError("argument to `keys` must be HASH, got %s",
					args[0].Type())
			}

			hashObj := args[0].(*Hash)
			newElements := make([]Object, len(hashObj.Pairs))
			var i int = 0

			for _, pair := range hashObj.Pairs {
				newElements[i] = pair.Key
				i += 1
			}

			return &Array{Elements: newElements}
		}},
	},
	{
		"delete", 
		&Builtin{Function: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments to `delete`. got=%d, want=2",
					len(args))
			}
			if args[0].Type() != HASH_OBJ {
				return newError("argument 1 to `delete` must be HASH, got %s",
					args[0].Type())
			}

			key, ok := args[1].(Hashable)
			if !ok {
				return newError("argument 2 to `delete` must be hashable, got %s",
					args[0].Type())
			}

			hashObj := args[0].(*Hash)

			hashKey := key.HashKey()
			delete(hashObj.Pairs, hashKey)
			
			return NULL
		}},
	},
	{
		"assign", 
		&Builtin{Function: func(args ...Object) Object {
			if len(args) != 3 {
				return newError("wrong number of arguments to `assign`. got=%d, want=3",
					len(args))
			}
			switch args[0].Type() {
			case ARRAY_OBJ:
				if args[1].Type() != INTEGER_OBJ {
					return newError("argument 2 to `assign` must be INTEGER, got %s",
						args[1].Type())
				}
				
				array := args[0].(*Array)
				length := int64(len(array.Elements))

				index := args[1].(*Integer).Value

				if (index > length - 1) {
					return newError("invalid index on array")
				}

				array.Elements[index] = args[2]
			
			case HASH_OBJ:
				hashKey, ok := args[1].(Hashable)
				if !ok {
					return newError("unusable as hash key: %s", args[1].Type())
				}

				hash := args[0].(*Hash)
				hash.Pairs[hashKey.HashKey()] = HashPair{Key: args[1], Value: args[2]}

			case STRING_OBJ:
				if args[1].Type() != INTEGER_OBJ {
					return newError("argument 2 to `assign` must be INTEGER, got %s",
						args[1].Type())
				}
				if args[2].Type() != STRING_OBJ {
					return newError("argument 3 to `assign` must be STRING, got %s",
						args[2].Type())
				}

				str := args[0].(*String)
				length := int64(len(str.Value))
				index := args[1].(*Integer).Value
				insert := args[2].(*String).Value

				if (index > length - 1) {
					return newError("invalid index on string")
				}

				str.Value = str.Value[:index] + insert + str.Value[index + 1:] 
			default:
				return newError("argument 1 to `assign` must be ARRAY, HASH, or STRING got %s",
					args[0].Type())
			}

			return NULL
		}},
	},
	{
		"type",
		&Builtin{Function: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments to `type`. got=%d, want=1",
					len(args))
			}
			
			return &String{Value: string(args[0].Type())}
		}},
	},
	{
		"command",
		&Builtin{Function: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments to `command`. got=%d, want=1",
					len(args))
			}
			str, ok := args[0].(*String)
			if !ok {
				return newError("argument to `command` must be STRING. got=%q",
					args[0].Type())
			}

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
				return newError("%s failed : %s\n", str.Value, err.Error())
			}

			stdout := &String{Value: outb.String()}
			stderr := &String{Value: errb.String()}

			stdoutKey := &String{Value: "stdout"}
			stdoutPair := HashPair{Key: stdoutKey, Value: stdout}

			stderrKey := &String{Value: "stderr"}
			stderrPair := HashPair{Key: stderrKey, Value: stderr}

			newHash := make(map[HashKey]HashPair)
			newHash[stdoutKey.HashKey()] = stdoutPair
			newHash[stderrKey.HashKey()] = stderrPair

			return &Hash{Pairs: newHash}
		}},
	},
	{
		"open",
		&Builtin{Function: func(args ...Object) Object {
			if len(args) > 2 || len(args) < 1{
				return newError("wrong number of arguments to `open`. got=%d, want=2",
					len(args))
			}
			if args[0].Type() != STRING_OBJ {
				return newError("argument 1 to `open` must be STRING. got=%q",
					args[0].Type())
			}
			var mode string
			if len(args) > 1 {
				if args[1].Type() != STRING_OBJ {
					return newError("argument 2 to `open` must be STRING. got=%q",
						args[1].Type())
				}

				mode = args[1].(*String).Value
			}

			var path string = args[0].(*String).Value

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

			fileObj := &File{Handle: file, Path: path}

			if md == os.O_RDONLY {
				fileObj.Reader = bufio.NewReader(file)
			} else {
				fileObj.Writer = bufio.NewWriter(file)
			}

			return fileObj
		}},
	},
	{
		"close",
		&Builtin{Function: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments to `close`. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != FILE_OBJ {
				return newError("argument to `close` must be FILE. got=%q",
					args[0].Type())
			}

			file := args[0].(*File).Handle
			file.Close()
			return TRUE
		}},
	},
	{
		"read",
		&Builtin{Function: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments to `read`. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != FILE_OBJ {
				return newError("argument to `read` must be FILE. got=%q",
					args[0].Type())
			}

			file := args[0].(*File)
			if file.Reader == nil {
				return &String{Value: ""}
			}

			out, err := file.Reader.ReadString('\n')
			if err != nil {
				return NULL
			}

			return &String{Value: out}
		}},
	},
	{
		"write",
		&Builtin{Function: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments to `write`. got=%d, want=2",
					len(args))
			}
			if args[0].Type() != FILE_OBJ {
				return newError("argument 1 to `write` must be FILE. got=%q",
					args[0].Type())
			}

			file := args[0].(*File)

			if file.Writer == nil {
				return FALSE
			}

			_, err := file.Writer.Write([]byte(args[1].Inspect()))
			if err == nil {
				file.Writer.Flush()
				return TRUE
			}

			return FALSE
		}},
	},
	{
		"remove",
		&Builtin{Function: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != FILE_OBJ {
				return newError("argument 1 to `remove` must be FILE. got=%q",
					args[0].Type())
			}

			file := args[0].(*File)

			err := os.Remove(file.Path)
			if err != nil {
				return &String{Value: err.Error()}
			}

			return TRUE
		}},
	},
	{
		"args",
		&Builtin{Function: func(args ...Object) Object {
			switch len(args) {
			case 0:
				length := len(os.Args[1:])
				out := make([]Object, length)
				for i, arg := range os.Args[1:] {
				out[i] = &String{Value: arg}
				}
				return &Array{Elements: out}
			case 1:
				if args[0].Type() != INTEGER_OBJ {
					return newError("argument to `args` must be INTEGER. got=%q",
						args[0].Type())
				}
				index := args[0].(*Integer).Value
				osArgs := os.Args[1:]
				if index > int64(len(osArgs)) - 1 || index < 0 {
					return newError("out of bounds index")
				}
				return &String{Value: osArgs[index]}
			default:
				return newError("wrong number of arguments. got=%d, want 0 or 1",
					len(args))
			}
		}},
	},
	{
		"wait",
		&Builtin{Function: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want 1",
					len(args))
			}
			switch args[0].Type() {
			case INTEGER_OBJ:
				period := args[0].(*Integer).Value
				time.Sleep(time.Duration(period) * time.Second)
			case FLOAT_OBJ:
				firstPeriod := args[0].(*Float).Value
				period := int64(1000 * firstPeriod)
				time.Sleep(time.Duration(period) * time.Millisecond)
			default:
				return newError("argument to `args` must be INTEGER or FLOAT. got=%q",
					args[0].Type())
			
			}
			return NULL
		}},
	},
	{
		"int",
		&Builtin{Function: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments to `int`. got=%d, want 1",
					len(args))
			}
			switch args[0].Type() {
			case FLOAT_OBJ:
				value := args[0].(*Float).Value
				return &Integer{Value: int64(value)}
			case STRING_OBJ:
				value := args[0].(*String).Value
				newValue, err := strconv.Atoi(value)
				if err != nil {
					return &Error{Message: err.Error()}
				}
				return &Integer{Value: int64(newValue)}
			default:
				return newError("argument to `int` must be FLOAT or STRING. got=%q",
					args[0].Type())
			}
		}},
	},
	{
		"float",
		&Builtin{Function: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments to `float. got%d, want 1",
					len(args))
			}
			switch args[0].Type() {
			case INTEGER_OBJ:
				value := args[0].(*Integer).Value
				return &Float{Value: float64(value)}
			case STRING_OBJ:
				value := args[0].(*String).Value
				newValue, err := strconv.ParseFloat(value, 64)
				if err != nil {
					return &Error{Message: err.Error()}
				}
				return &Float{Value: newValue}
			default:
				return newError("argument to `float` must be INTEGER or STRING. got=%q",
					args[0].Type())
			}
		}},
	},
	{
		"rand",
		&Builtin{Function: func(args ...Object) Object {
			return &Float{Value: rand.Float64()}
		}},
	},
}

func newError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

func GetBuiltinByName(name string) *Builtin {
	for _, def := range Builtins {
		if def.Name == name {
			return def.Builtin
		}
	}
	return nil
}
