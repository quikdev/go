package commands

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/quikdev/go/context"
	"github.com/quikdev/go/util"
)

type Kill struct {
	Profile []string `name:"profile" optional:"" help:"Name of the manifest.json profile attribute to apply."`
	Args    []string `arg:"" optional:"" help:"Names of the executables to kill (ex: myapp.exe/myapp)"`
}

func (k *Kill) Run(c *Context) error {
	ctx := context.New(k.Profile...)
	ctx.Configure()

	if len(k.Args) == 0 {
		out := ctx.OutputFile()

		if len(out) == 0 {
			util.Stderr("please specify a process to kill (ex: myapp.exe or someapp)", true)
		}

		k.Args = []string{out}
	}

	for _, binary := range k.Args {
		pids, err := findProcessIDs(binary)
		if err != nil {
			util.Stderr(fmt.Sprintf("Failed to find processes: %v", err))
		} else {
			if len(pids) == 0 {
				util.Stdout(fmt.Sprintf("No running %s processes", binary))
			} else {
				if err := killProcesses(pids); err != nil {
					util.Stderr(fmt.Sprintf("failed to kill process(es): %v\n", err))
				} else {
					util.Stdout(fmt.Sprintf("\nSuccessfully killed %d %s process(es)\n", len(pids), binary))
				}
			}
		}
	}

	return nil
}

func killProcesses(pids []string) error {
	if len(pids) == 0 {
		return nil
	}

	if runtime.GOOS == "windows" {
		args := []string{"/F"}
		for _, pid := range pids {
			args = append(args, fmt.Sprintf("/PID %s", pid))
		}
		return util.Stream("taskkill " + strings.Join(args, " "))
	}

	args := append([]string{"-9"}, pids...)
	return util.Stream("kill " + strings.Join(args, " "))
}

func findProcessIDs(binaryName string) ([]string, error) {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("tasklist", "/FI", fmt.Sprintf("IMAGENAME eq %s", binaryName))
	} else {
		cmd = exec.Command("pgrep", "-f", binaryName)
	}

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	var processIDs []string
	if runtime.GOOS == "windows" {
		lines := strings.Split(out.String(), "\n")
		for _, line := range lines {
			if strings.Contains(line, binaryName) {
				fields := strings.Fields(line)
				if len(fields) > 1 {
					processIDs = append(processIDs, fields[1])
				}
			}
		}
	} else {
		processIDs = strings.Split(strings.TrimSpace(out.String()), "\n")
	}

	return processIDs, nil
}
