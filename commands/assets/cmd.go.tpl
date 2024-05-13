package main

import (
	"flag"
	"fmt"
	"os"
)

var (
  name string
  version string
  description string
  buildTime string
)

func init() {
  fmt.Println("(Created on %v)\n", buildTime)
}

// TODO: Write my code here
func main() {
	// Define command-line flags
	person := flag.String("person", "World", "Specify a name")
	age := flag.Int("age", 0, "Specify an age")
	showVersion := flag.Bool("v", false, "Show version")

	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	// Check if age is provided
	if *age == 0 {
		fmt.Println("Please provide your age using -age flag")
		os.Exit(1)
	}

	// Print out the greeting with the provided name and age
  fmt.Printf("%s are %d years old.\n", person, *age)
}
