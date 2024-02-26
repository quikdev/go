package util

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

func Stdout(txt string, exit ...bool) {
	stdout := color.New(color.Faint).SprintFunc()
	fmt.Println(stdout(txt))

	if len(exit) > 0 {
		if exit[0] {
			os.Exit(0)
		}
	}
}

func Stderr(msg interface{}, exit ...bool) {
	stderr := color.New(color.FgRed).SprintFunc()

	if _, ok := msg.(error); ok {
		fmt.Println(stderr(msg.(error).Error()))
	} else {
		fmt.Println(stderr(msg.(string)))
	}

	if len(exit) > 0 {
		if exit[0] {
			os.Exit(1)
		}
	}
}

func command(cmd ...string) string {
	c := color.New(color.FgCyan).SprintFunc()
	return c(strings.Join(cmd, " "))
}

func Command(cmd ...string) {
	fmt.Println(command(cmd...))
}

func HighlightComment(txt string) {
	colorize := color.New(color.FgMagenta, color.Faint).SprintFunc()
	fmt.Println(colorize("# " + txt))
}

func HighlightCommand(cmd string, args ...string) {
	dim := color.New(color.FgYellow, color.Faint).SprintFunc()
	c := command(cmd)

	if len(args) > 0 {
		for _, arg := range args {
			c = c + " " + dim(arg)
		}
	}

	fmt.Println(c)
}

func Highlight(text ...string) {
	c := color.New(color.FgYellow, color.Bold).SprintFunc()
	fmt.Println(c(strings.Join(text, " ")))
}

func SubtleHighlight(text ...string) {
	c := color.New(color.FgYellow, color.Faint).SprintFunc()
	fmt.Println(c(strings.Join(text, " ")))
}
