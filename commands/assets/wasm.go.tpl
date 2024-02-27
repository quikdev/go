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
	// "Export" the function for JavaScript
	js.Global().Set("hello", hello())

	// Indicate the WASM is recognized by JS
	fmt.Println("Go WebAssembly recognized!")

	// Prevent the WASM from exiting prematurely
	<-make(chan struct{})
}

func sayhello() string {
	return fmt.Sprintf("Hello, WebAssembly! From %v v%v (%v).", name, version, description)
}

// WASM function (js.Func)
func hello() js.Func {
	fn := js.FuncOf(func(this js.Value, args []js.Value) any {
		return sayhello()
	})

	return fn
}
