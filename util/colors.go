package util

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

func Stdout(txt string, exit ...bool) {
	stdout := color.New(color.Faint).SprintFunc()
	fmt.Printf("%v", stdout(txt))

	if len(exit) > 0 {
		if exit[0] {
			os.Exit(0)
		}
	}
}

func Stderr(msg interface{}, exit ...bool) {
	stderr := color.New(color.FgRed).SprintFunc()

	if _, ok := msg.(error); ok {
		fmt.Printf("%v", stderr(msg.(error).Error()))
	} else {
		fmt.Printf("%v", stderr(msg.(string)))
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
	fmt.Println(Highlighter(text...))
}

func Highlighter(text ...string) string {
	c := color.New(color.FgYellow, color.Bold).SprintFunc()
	return fmt.Sprintln(c(strings.Join(text, " ")))
}

func SubtleHighlight(text ...string) {
	fmt.Println(SubtleHighlighter(text...))
}

func SubtleHighlighter(text ...string) string {
	c := color.New(color.FgYellow, color.Faint).SprintFunc()
	return fmt.Sprintln(c(strings.Join(text, " ")))
}

func Dim(text ...string) string {
	c := color.New(color.Faint).SprintFunc()
	return c(strings.Join(text, " "))
}
