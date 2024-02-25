package main

import (
	"fmt"
	"os"

	"github.com/quikdev/go/commands"
	"github.com/quikdev/go/util"

	"github.com/alecthomas/kong"
)

var (
	name        string
	description string
	version     string
)

func main() {
	debug := false
	val := os.Getenv("DEBUG")
	if val == "true" {
		debug = true
	}

	cmd := &commands.Context{
		Debug: debug,
	}

	if len(os.Args) < 2 {
		os.Args = append(os.Args, "--help")
	} else if len(os.Args) >= 2 && (util.InSlice[string]("-v", os.Args) || util.InSlice[string]("--version", os.Args)) {
		fmt.Println(version)
		return
	}

	root := &commands.Root

	ctx := kong.Parse(
		root,
		kong.Name(name),
		kong.Description(description+"\nv"+version),
		kong.UsageOnError(),
	)

	ctx.Run(cmd)
}
