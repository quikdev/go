package commands

import "github.com/alecthomas/kong"

var Root struct {
	Init      Init             `cmd:"init" short:"i" help:"Setup a new Go module or application"`
	Build     Build            `cmd:"build" short:"b" help:"Build the Go application"`
	Run       Run              `cmd:"run" short:"r" help:"Run the Go application"`
	Test      Test             `cmd:"test" short:"t" help:"Run unit tests"`
	Uninstall Uninstall        `cmd:"uninstall" short:"u" help:"Uninstall a 'go install' app."`
	Exec      Do               `cmd:"exec" short:"x" help:"Run a script from the manifest"`
	Bump      Bump             `cmd:"bump" help:"Bump the semantic version number in the manifest"`
	Todo      Todo             `cmd:"todo" help:"List all of the todo items found in the code base."`
	Kill      Kill             `cmd:"kill" short:"k" help:"Kill processes by executable name."`
	Version   kong.VersionFlag `name:"version" short:"v" help:"Display the QuikGo version."`
}
