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

	"github.com/quikdev/go/util"
)

type Command struct {
	str   []string
	color map[string][]int
}

func New() *Command {
	return &Command{str: []string{}, color: make(map[string][]int)}
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
	return colorize(cmd.String(), help...)
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

func (cmd *Command) Run(cwd ...string) {
	commands := split(cmd.str, "&&")

	for _, code := range commands {
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

		var wg sync.WaitGroup

		// Function to read and print output from a pipe
		stream := func(pipe io.Reader) {
			defer wg.Done()
			scanner := bufio.NewScanner(pipe)
			for scanner.Scan() {
				fmt.Println(scanner.Text())
			}
		}

		wg.Add(2)
		go stream(stdout)
		go stream(stderr)

		// Wait for the command to finish
		c.Wait()
		wg.Wait()
	}
}
