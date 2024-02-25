package commands

import "github.com/alecthomas/kong"

var Root struct {
	Init    Init             `cmd:"init" short:"i" help:"Initialize a new Go application"`
	Build   Build            `cmd:"build" short:"b" help:"Build the Go application"`
	Run     Run              `cmd:"run" short:"r" help:"Run the Go application"`
	Test    Test             `cmd:"test" short:"t" help:"Run unit tests"`
	Version kong.VersionFlag `name:"version" short:"v" help:"Display the application version."`
}
