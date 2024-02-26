// main.go
package main

import (
	"fmt"
	"syscall/js"
)

var (
	name        string
	description string
	version     string
)

func main() {
	js.Global().Set("hello", hello())
	fmt.Println("Go WebAssembly recognized!")
	<-make(chan struct{})
}

func sayhello() string {
	return fmt.Sprintf("Hello, WebAssembly! From %v v%v (%v).", name, version, description)
}

func hello() js.Func {
	fn := js.FuncOf(func(this js.Value, args []js.Value) any {
		return sayhello()
	})

	return fn
}
