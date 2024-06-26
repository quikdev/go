package command

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/quikdev/go/config"
	"github.com/quikdev/go/util"
)

type InjectedCommand struct {
	Position int
	Command  string
	Noerror  bool
	Before   bool
}

type Command struct {
	str      []string
	color    map[string][]int
	pid      int
	injected []InjectedCommand
}

func New() *Command {
	return &Command{
		str:      []string{},
		color:    make(map[string][]int),
		injected: []InjectedCommand{},
	}
}

func (cmd *Command) InjectCommand(position int, command string, before bool, noerr ...bool) {
	noerror := false
	if len(noerr) > 0 {
		noerror = noerr[0]
	}

	cmd.injected = append(cmd.injected, InjectedCommand{
		Position: position,
		Command:  command,
		Noerror:  noerror,
		Before:   before,
	})
}

func (cmd *Command) Add(value ...string) {
	cmd.str = append(cmd.str, value...)
}

func (cmd *Command) String(includeApp ...bool) string {
	include := true
	if len(includeApp) > 0 {
		include = includeApp[0]
	}

	if !include {
		return strings.Join(cmd.str[1:], " ")
	}

	return strings.Join(cmd.str, " ")
}

func (cmd *Command) Display(help ...bool) string {
	command := colorize(cmd.String(), help...)
	return command
}

func (cmd *Command) Kill() error {
	process, err := os.FindProcess(cmd.pid)
	if err != nil {
		return err
	}
	return process.Kill()
}

func (cmd *Command) PID() int {
	return cmd.pid
}

func (cmd *Command) Run(cwd ...string) {
	commands := split(cmd.str, "&&")

	cfg := config.New()
	vars := cfg.GetEnvVarList()

	for index, code := range commands {
		cmd.runInjectedCommand(index, true)

		args := make([]string, len(code))
		for i, line := range code {
			if i > 1 {
				if code[i-1] == "-ldflags" || code[i-1] == "-gccgoflags" || code[i-1] == "-gcflags" || code[i-1] == "-asmflags" {
					if strings.HasPrefix(line, "\"") && strings.HasSuffix(line, "\"") {
						line = line[1 : len(line)-1]
					}
				}
			}
			args[i] = line
		}

		for i := range args {
			args[i] = filepath.ToSlash(args[i])
		}

		c := exec.Command(code[0], args[1:]...)

		// Add manifest/package environment variables to command execution context
		c.Env = append(os.Environ(), vars...)

		curr, _ := os.Getwd()
		if len(cwd) > 0 {
			if cwd[0] != "./" && cwd[0] != curr {
				util.Stdout(fmt.Sprintf("# using \"%s\" as working directory...\n", cwd[0]))
			}
			c.Dir = cwd[0]
		}

		stdout, err := c.StdoutPipe()
		if err != nil {
			fmt.Printf("Error creating stdout pipe: %v\n", err)
			return
		}

		stderr, err := c.StderrPipe()
		if err != nil {
			fmt.Printf("Error creating stderr pipe: %v\n", err)
			return
		}

		if err := c.Start(); err != nil {
			fmt.Printf("Error starting command: %v\n", err)
			return
		}

		cmd.pid = c.Process.Pid

		var wg sync.WaitGroup

		// Function to read and print output from a pipe
		exit := false
		stream := func(pipe io.Reader, streamType string) {
			defer wg.Done()
			size := 0
			scanner := bufio.NewScanner(pipe)
			for scanner.Scan() {
				txt := scanner.Text()
				fmt.Println(txt)
				size += len(txt)
			}

			if streamType == "stderr" && size > 0 {
				exit = true
			}
		}

		wg.Add(2)
		go stream(stdout, "stdout")
		go stream(stderr, "stderr")

		// Wait for the command to finish
		c.Wait()
		wg.Wait()

		cmd.runInjectedCommand(index, false)

		if exit {
			os.Exit(1)
		}
	}
}

// func (cmd *Command) displayInjectedCommand(index int, command string) string {
// 	for _, ic := range cmd.injected {
// 		if ic.Position == index {
// 			if ic.Before {
// 				command = "\n&& " + ic.Command + "|" + command
// 			} else {
// 				command += "\n&& " + ic.Command
// 			}
// 		}
// 	}

// 	return strings.ReplaceAll(strings.ReplaceAll(command, "|\n&&", "\n&&"), "|", "\n&&")
// }

func (cmd *Command) runInjectedCommand(index int, before bool) {
	for _, ic := range cmd.injected {
		if ic.Position == index && ic.Before == before {
			if ic.Noerror {
				util.BailOnError(util.StreamNoStdErr(ic.Command))
			} else {
				util.BailOnError(util.Stream(ic.Command))
			}
			fmt.Println("")
		}
	}
}

func split(main []string, delimiter string) [][]string {
	var result [][]string
	var smallerSlice []string

	for _, item := range main {
		if item == delimiter {
			// Append the current smaller slice to the result if it's not empty
			if len(smallerSlice) > 0 {
				result = append(result, smallerSlice)
				smallerSlice = nil
			}
		} else {
			// Add the item to the current smaller slice
			smallerSlice = append(smallerSlice, item)
		}
	}

	// Append the last smaller slice to the result
	if len(smallerSlice) > 0 {
		result = append(result, smallerSlice)
	}

	return result
}
