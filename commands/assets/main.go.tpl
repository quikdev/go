package main

import (
  "fmt"
)

var (
  name string
  version string
  description string
  buildTime string
)

// TODO: Write my code
func main() {
  fmt.Println(SayHello())
}

func SayHello() string {
  return fmt.Sprintf("Name: %s\nVersion: %v\nDescription: %v\n", name, version, description)
}
