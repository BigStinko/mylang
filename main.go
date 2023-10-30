package main

import (
	"fmt"
	"os"
	"os/user"
	"mylang/repl"
)
func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Hello %s, this is mylang\n", user.Username)
	repl.StartREPL(os.Stdin, os.Stdout)
}
