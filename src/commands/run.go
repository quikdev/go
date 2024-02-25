package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/quikdev/go/context"

	fs "github.com/coreybutler/go-fsutil"
)

type Run struct {
	Bundle []string `name:"bundle" short:"b" type:"string" help:"Bundle the application into a tarball/zipball" enum:"zip,tar"`
	OS     []string `name:"os" type:"string" help:"The operating system(s) to build for (any options from 'go tool dist list' is valid)"`
	WASM   bool     `name:"wasm" type:"bool" help:"Output a web assembly (OS is ignored when this option is true)"`
	Output string   `name:"output" short:"o" type:"string" help:"Output file name"`
	Tips   bool     `name:"tips" short:"t" type:"bool" help:"Display tips in the generated commands"`
	Minify bool     `name:"minify" short:"m" type:"bool" help:"Set ldflags to strip debugging symbols and remove DWARF generations"`
	Shrink bool     `name:"shrink" short:"s" type:"bool" help:"Set gccgoflags to strip debugging symbols and remove DWARF generations"`
	DryRun bool     `name:"dry-run" short:"d" type:"bool" help:"Display the command without executing it."`
	File   string   `arg:"source" optional:"" help:"Go source file (ex: main.go)"`
	// Container string `name:"container" default:"docker" type:"string" enum:"docker,podman" help:"The containerization technology to build with"`
}

func (b *Run) Run(c *Context) error {
	ctx := context.New()
	ctx.Configure()

	if b.WASM {
		ctx.OS = []string{"js"}
	} else {
		ctx.OS = b.OS
	}

	if b.Minify {
		ctx.StripSymbols = true
		ctx.StripDebugging = true
	}

	if b.Shrink {
		ctx.GCCGoFlags.Add("-s")
		ctx.GCCGoFlags.Add("-w")
	}

	cmd := ctx.RunCommand()
	if c.Debug {
		b.Tips = true
	}

	// Display command
	fmt.Println(cmd.Display(b.Tips))

	parentdir := filepath.Dir(ctx.Output())
	if !fs.Exists(parentdir) {
		os.MkdirAll(parentdir, os.ModePerm)
	}

	// Run command
	if !b.DryRun {
		cmd.Run(ctx.CWD)
	}

	return nil
}
