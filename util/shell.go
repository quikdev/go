package util

import (
	"os"
	"os/exec"
	"strings"
)

func Output(cmd string, cwd ...string) ([]byte, error) {
	Highlight(cmd)

	var wd string
	if len(cwd) > 0 {
		// command.Dir = cwd[0]
		wd, _ = os.Getwd()
		err := os.Chdir(cwd[0])
		if err != nil {
			return []byte{}, err
		}
	}

	parts := strings.Split(cmd, " ")
	command := exec.Command(parts[0], parts[1:]...)

	out, err := command.Output()
	if len(cwd) > 0 {
		os.Chdir(wd)
	}

	if err != nil {
		Stderr(err)
	}

	if len(out) > 0 {
		Stdout(string(out))
	}

	return out, err
}

func Run(cmd string, cwd ...string) error {
	_, err := Output(cmd, cwd...)
	return err
}
