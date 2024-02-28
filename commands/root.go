package commands

import "github.com/alecthomas/kong"

var Root struct {
	Init      Init             `cmd:"init" short:"i" help:"Setup a new Go module or application"`
	Build     Build            `cmd:"build" short:"b" help:"Build the Go application"`
	Run       Run              `cmd:"run" short:"r" help:"Run the Go application"`
	Test      Test             `cmd:"test" short:"t" help:"Run unit tests"`
	Uninstall Uninstall        `cmd:"uninstall" short:"u" help:"Uninstall a 'go install' app."`
	Version   kong.VersionFlag `name:"version" short:"v" help:"Display the QuikGo version."`
}
