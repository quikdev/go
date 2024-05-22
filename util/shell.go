package util

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
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

		defer func() {
			os.Chdir(wd)
		}()
	}

	parts := strings.Split(cmd, " ")
	command := exec.Command(parts[0], parts[1:]...)

	out, err := command.Output()

	if err != nil {
		Stderr(err)
	}

	if len(out) > 0 {
		Stdout(string(out))
	}

	return out, err
}

func streamcmd(raw bool, cmd string, cwd ...string) error {
	Highlight(cmd)

	return streamcmdnohighlight(raw, cmd, false, cwd...)
}

func streamcmdnohighlight(raw bool, cmd string, nostderr bool, cwd ...string) error {
	// var wg sync.WaitGroup
	var wd string
	if len(cwd) > 0 {
		wd, _ = os.Getwd()
		err := os.Chdir(cwd[0])
		if err != nil {
			return err
		}

		defer func() {
			os.Chdir(wd)
		}()
	}

	if runtime.GOOS == "windows" && !strings.HasPrefix(cmd, "cmd") {
		cmd = "cmd /c " + cmd
	}

	parts := strings.Split(cmd, " ")
	command := exec.Command(parts[0], parts[1:]...)

	stdout, err := command.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := command.StderrPipe()
	if err != nil {
		return err
	}

	if err := command.Start(); err != nil {
		return err
	}

	// wg.Add(2)

	var errMsg string
	go stream(raw, stdout, "stdout")
	if nostderr {
		go stream(raw, stderr, "stdout")
	} else {
		go stream(raw, stderr, "stderr", &errMsg)
	}

	if err := command.Wait(); err != nil {
		return err
	}

	// wg.Wait()

	if len(errMsg) > 0 {
		return errors.New(errMsg)
	}

	return nil
}

func Stream(cmd string, cwd ...string) error {
	return streamcmd(false, cmd, cwd...)
}

func StreamNoStdErr(cmd string, cwd ...string) error {
	Highlight(cmd)

	return streamcmdnohighlight(false, cmd, true, cwd...)
}

func StreamNoHighlight(cmd string, cwd ...string) error {
	return streamcmdnohighlight(false, cmd, false, cwd...)
}

func StreamNoStdErrNoHighlight(cmd string, cwd ...string) error {
	return streamcmdnohighlight(true, cmd, true, cwd...)
}

func StreamRaw(cmd string, cwd ...string) error {
	return streamcmd(true, cmd, cwd...)
}

func StreamRawNoHighlight(cmd string, cwd ...string) error {
	return streamcmdnohighlight(true, cmd, false, cwd...)
}

func stream(raw bool, reader io.Reader, streamType string /*, wg *sync.WaitGroup*/, errMsg ...*string) {
	// defer wg.Done()

	// Create a buffer for reading from the reader
	buf := make([]byte, 1024)
	for {
		// Read from the reader
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		// Print the output to console
		out := string(buf[:n])

		if streamType == "stderr" {
			if raw {
				fmt.Print(out)
			} else {
				Stderr(out)
			}
			if len(errMsg) > 0 {
				o := errMsg[0]
				*o += out
			}
		} else {
			if raw {
				fmt.Print(out)
			} else {
				Stdout(out)
			}
		}
	}
}

func Run(cmd string, cwd ...string) error {
	_, err := Output(cmd, cwd...)
	return err
}
