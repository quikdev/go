package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	tap "github.com/mpontillo/tap13"
	"github.com/quikdev/go/v1/context"
	"github.com/quikdev/go/v1/spec"
	"github.com/quikdev/go/v1/util"
)

var (
	total       int
	emptyString string
	emptyInt    int
	lastTest    string
	tapversion  int
	result      string
)

type Test struct {
	Format string `name:"format" short:"f" default:"spec" help:"The format to diplay test results in. Defaults to 'spec', a TAP visualizer. Options include 'tap', 'spec', 'json', and 'go' (i.e. go test standard)"`
}

func (t *Test) Run(c *Context) error {
	ctx := context.New()
	ctx.Configure()

	format := strings.ToLower(t.Format)

	args := []string{"test", "./..."}
	if format != "go" {
		args = append(args, "-json")
	}

	cmd := exec.Command("go", args...)

	stdout, err := cmd.StdoutPipe()
	util.BailOnError(err)
	err = cmd.Start()
	util.BailOnError(err)

	if format == "tap" || format == "tap13" || format == "tap14" || format == "spec" {
		tapversion = 14
		if format == "tap13" {
			tapversion = 13
		}
		if format != "spec" {
			fmt.Printf("TAP version %d\n", tapversion)
		}
	}

	var wg sync.WaitGroup
	pr, pw := io.Pipe()

	wg.Add(2)

	go func() {
		defer wg.Done()
		process(format, bufio.NewScanner(stdout), pw)
	}()

	go func() {
		defer wg.Done()
		formatresult(format, pr)
	}()

	wg.Wait()

	return nil
}

func process(format string, scanner *bufio.Scanner, writer *io.PipeWriter) {
	defer writer.Close()

	if format == "spec" {
		writer.Write([]byte("TAP version 14\n"))
	}

	for scanner.Scan() {
		line := scanner.Bytes()
		_, err := writer.Write(append(line, byte('\n')))
		if err != nil {
			util.Stderr(err)
			break
		}
	}

	err := scanner.Err()
	if err != nil && err != io.EOF {
		util.Stderr(err, true)
	}
}

func formatresult(format string, reader io.Reader) {
	formatter := spec.Formatter()
	scanner := bufio.NewScanner(reader)
	collecting := false
	lines := []string{}
	var tapline string
	spec := []string{"TAP version 14"}

	for scanner.Scan() {
		line := scanner.Text()
		switch format {
		case "json":
			fallthrough
		case "go":
			fmt.Println(line)
		default:
			if strings.Contains(line, "\"Test\":") {
				if strings.Contains(line, "\"Action\":\"run\"") {
					collecting = true
					lines = []string{}
				}
				if strings.Contains(line, "\"Action\":\"pass\"") || strings.Contains(line, "\"Action\":\"fail\"") || strings.Contains(line, "\"Action\":\"skip\"") {
					collecting = false
				}
				if collecting {
					lines = append(lines, line)
				} else {
					tapline, lastTest = tapFormat(strings.Join(lines, "\n"), lastTest)
					if format == "spec" {
						spec = append(spec, tapline)
					} else if len(tapline) > 0 {
						fmt.Println(tapline)
					}
					collecting = false
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		util.Stderr(err, true)
	}

	if strings.HasPrefix(format, "tap") {
		fmt.Printf("1..%d\n", total)
	}

	if format == "spec" {
		result := tap.Parse(strings.Split(strings.Join(spec, "\n"), "\n"))
		formatter.Format(result)
		fmt.Println("")
		formatter.Summary()
	}
}

func tapFormat(input string, last string) (string, string) {
	out := TapOutput{Diagnostic: []string{}, Passed: true}

	chunks := strings.Split(strings.TrimSpace(input), "\n")
	for _, chunk := range chunks {
		var result TestResult

		if err := json.Unmarshal([]byte(chunk), &result); err != nil {
			return "", last
		}

		switch strings.ToLower(result.Action) {
		case "run":
			out.Label = result.Package
			out.Test = result.Test
		case "pass":
			out.Passed = true
		case "fail":
			out.Passed = false
		case "output":
			if !strings.HasPrefix(result.Output, "===") && !strings.HasPrefix(result.Output, "---") && !strings.HasPrefix(result.Output, "FAIL") {
				out.Diagnostic = append(out.Diagnostic, strings.TrimSpace(result.Output))
			}
		}
	}

	total += 1
	out.TestNumber = total

	return out.Format(), last
}
