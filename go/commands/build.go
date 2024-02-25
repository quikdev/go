package commands

import (
	"fmt"
	"path/filepath"

	"github.com/quikdev/go/context"
	"github.com/quikdev/go/util"

	goupx "github.com/alegrey91/go-upx"
	fs "github.com/coreybutler/go-fsutil"
	"github.com/dustin/go-humanize"
)

type Build struct {
	Bundle   []string `name:"bundle" short:"b" type:"string" help:"Bundle the application into a tarball/zipball" enum:"zip,tar"`
	OS       []string `name:"os" type:"string" help:"The operating system(s) to build for (any options from 'go tool dist list' is valid)"`
	WASM     bool     `name:"wasm" type:"bool" help:"Output a web assembly (OS is ignored when this option is true)"`
	Output   string   `name:"output" short:"o" type:"string" help:"Output file name"`
	Tips     bool     `name:"tips" short:"t" type:"bool" help:"Display tips in the generated commands"`
	Minify   bool     `name:"minify" short:"m" type:"bool" help:"Set ldflags to strip debugging symbols and remove DWARF generations"`
	Shrink   bool     `name:"shrink" short:"s" type:"bool" help:"Set gccgoflags to strip debugging symbols and remove DWARF generations"`
	Compress bool     `name:"compress" short:"c" type:"bool" help:"Compress with UPX"`
	DryRun   bool     `name:"dry-run" short:"d" type:"bool" help:"Display the command without executing it."`
	File     string   `arg:"source" optional:"" help:"Go source file (ex: main.go)"`
	// Container string `name:"container" default:"docker" type:"string" enum:"docker,podman" help:"The containerization technology to build with"`
}

func (b *Build) Run(c *Context) error {
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

	cmd := ctx.BuildCommand()
	if c.Debug {
		b.Tips = true
	}

	ctx.OutputPath = "./bin"

	// Display command
	fmt.Println(cmd.Display(b.Tips))

	parentdir := filepath.Dir(ctx.Output())
	if !fs.Exists(parentdir) {
		fs.Mkdirp(parentdir)
	}

	// Run command
	if !b.DryRun {
		cmd.Run(ctx.CWD)

		if b.Compress {
			util.Stdout("\n# compressing executable\n")
			upx := goupx.NewUPX()
			options := goupx.Options{
				CompressionTuningOpt: goupx.CompressionTuningOptions{
					Brute: 1,
				},
			}
			_, err := upx.Compress(ctx.Output(), 9, options)
			util.BailOnError(err)
			util.HighlightCommand("upx", upx.GetArgs()...)
			util.Stdout(fmt.Sprintf("  ↳ Name: %v\n  ↳ Format: %v\n  ↳ Original: %v\n  ↳ Compressed: %v\n  ↳ Compression Ratio: %v\n", upx.CmdExecution.GetName(), upx.CmdExecution.GetFormat(), humanize.Bytes(upx.CmdExecution.GetOriginalFileSize()), humanize.Bytes(upx.CmdExecution.GetCompressedFileSize()), upx.CmdExecution.GetRatio()))
		}
	}

	return nil
}
