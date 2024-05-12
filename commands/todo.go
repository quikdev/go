package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Todo struct {
	Format string `name:"format" short:"f" default:"text" help:"output in a specific format" enum:"text,md,markdown,json"`
	Save   bool   `name:"save" short:"s" help:"save output to disk"`
	File   string `name:"output" short:"o" help:"output file name/location" default:"./todos.txt"`
}

type Item struct {
	Path    string
	Line    int
	Comment string
}

func (i *Item) Text() string {
	return fmt.Sprintf(`%s
  at %s:%d`, i.Comment, i.Path, i.Line)
}

func (i *Item) Markdown() string {
	return fmt.Sprintf(`[ ] %s _at %s:**%d**_`, i.Comment, i.Path, i.Line)
}

func (t *Todo) Run(c *Context) error {
	input := "./"

	// ctx := context.New()
	// ctx.Configure()
	// cfg := ctx.GetConfig()

	var items []*Item

	err := filepath.Walk(input, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".go" {
			xerr := extractTODOs(path, &items)
			if xerr != nil {
				return xerr
			}
		}

		return nil
	})

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	format := strings.ToLower(t.Format)
	out := []string{}
	for _, item := range items {
		switch format {
		case "markdown":
			fallthrough
		case "md":
			out = append(out, item.Markdown())
		default:
			out = append(out, item.Text())
		}
	}

	cr := "\n\n"

	switch format {
	case "json":
		data, _ := json.MarshalIndent(items, "", "  ")
		println(string(data))
	case "markdown":
		fallthrough
	case "md":
		cr = "\n"
		fallthrough
	default:
		if len(out) == 0 {
			println("no TODO's!")
		} else {
			println(strings.Join(out, cr))
		}
	}

	return nil
}

func extractTODOs(filePath string, items *[]*Item) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var multilineComment strings.Builder
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "/*") {
			for strings.TrimSpace(line); !strings.HasSuffix(line, "*/") && scanner.Scan(); {
				lineNumber++
				line = scanner.Text()
			}
			continue
		}

		if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") {
			lowerLine := strings.ToLower(line)
			if strings.Contains(lowerLine, "todo:") || strings.Contains(lowerLine, "todo -") {
				todoIndex := strings.IndexAny(lowerLine, ":")
				if todoIndex == -1 {
					todoIndex = strings.Index(lowerLine, "-")
				}
				comment := strings.TrimSpace(line[todoIndex+1:])
				multilineComment.WriteString(comment)
				for scanner.Scan() {
					lineNumber++
					line = scanner.Text()
					line = strings.TrimSpace(line)
					if line == "" || (!strings.HasPrefix(line, "//") && !strings.HasPrefix(line, "/*")) {
						*items = append(*items, &Item{Path: filePath, Line: lineNumber, Comment: multilineComment.String()})
						multilineComment.Reset()
						break
					}
					if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") {
						lowerLine := strings.ToLower(line)
						if strings.Contains(lowerLine, "todo:") || strings.Contains(lowerLine, "todo -") {
							todoIndex := strings.IndexAny(lowerLine, ":")
							if todoIndex == -1 {
								todoIndex = strings.Index(lowerLine, "-")
							}
							if multilineComment.Len() > 0 {
								*items = append(*items, &Item{Path: filePath, Line: lineNumber, Comment: multilineComment.String()})
								multilineComment.Reset()
							}
							comment := strings.TrimSpace(line[todoIndex+1:])
							multilineComment.WriteString(comment)
						}
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
