package commands

import (
	"fmt"
	"strings"

	"github.com/quikdev/go/context"
	"github.com/quikdev/go/util"
)

type Do struct {
	Script string   `arg:"script" help:"Name of the script to execute"`
	Args   []string `arg:"arguments" optional:"" help:"Additional arguments to pass to the script"`
}

func (d *Do) Run(c *Context) error {
	ctx := context.New()
	ctx.Configure()

	scripts, exists := ctx.GetConfig().Get("scripts")
	if !exists {
		util.Stderr("no scripts defined", true)
	}

	if cmd, exists := scripts.(map[string]interface{})[d.Script]; exists {
		command := []string{cmd.(string)}
		if len(d.Args) > 0 {
			// command = append(command, "--")
			command = append(command, d.Args...)
		}
		util.StreamRawNoHighlight(strings.Join(command, " "))
	} else {
		util.Stderr(fmt.Sprintf(`"%s" script does not exist or cannot be found`, d.Script))
	}

	return nil
}
