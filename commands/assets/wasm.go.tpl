// main.go
package main

import (
  "fmt"
)

var (
  name string
  description string
  version string
)

func main() {
  fmt.Printf("Hello, WebAssembly! From %vv%v (%v).", name, version, description)
}
