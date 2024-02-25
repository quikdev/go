package util

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"html/template"
	"log"
	"os"
	"path/filepath"

	fs "github.com/coreybutler/go-fsutil"
)

func FileExists(name string) bool {
	return fs.Exists(name)
}

func ReadTextFile(path string) (string, error) {
	return fs.ReadTextFile(path)
}

func FindGoWorkFile(path string) (string, bool) {
	// Check if the path is empty
	if path == "" {
		return "", false
	}

	// Get the absolute path of the input
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", false
	}

	// Iterate over parent directories until volume root is reached
	for {
		fullPath := filepath.Join(absPath, "go.work")

		// Check if the file exists
		_, err := os.Stat(fullPath)
		if err == nil {
			// File found
			return fullPath, true
		} else if !os.IsNotExist(err) {
			// Error occurred (other than file not found)
			return "", false
		}

		// Move to the parent directory
		parent := filepath.Dir(absPath)
		if parent == absPath {
			// Reached the volume root
			break
		}
		absPath = parent
	}

	// File not found up to the volume root
	return "", false
}

func FindMainFileInDirectory(directoryPath string) (string, error) {
	// Create a list of .go files in the specified directory
	goFiles := []string{}
	err := filepath.Walk(directoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".go" {
			goFiles = append(goFiles, path)
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	// Iterate over .go files and check if they contain a main() function
	for _, goFile := range goFiles {
		fileSet := token.NewFileSet()
		fileNode, err := parser.ParseFile(fileSet, goFile, nil, parser.DeclarationErrors)
		if err != nil {
			return "", err
		}

		// Check if the file contains a main() function
		for _, decl := range fileNode.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
				return goFile, nil
			}
		}
	}

	return "", fmt.Errorf("No .go file with main() function found in the directory")
}

func WriteTextFile(path string, content string, exitOnError ...bool) error {
	return WriteFile(path, []byte(content), exitOnError...)
}

func WriteFile(path string, content []byte, exitOnError ...bool) error {
	file, err := os.Create(path)
	if err != nil {
		Stderr(err, exitOnError...)
		return err
	}

	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		Stderr(err, exitOnError...)
		return err
	}

	Stdout("  ↳ created " + path)

	return nil
}

func ApplyTemplate(tpl []byte, data interface{}) string {
	t, err := template.New("").Parse(string(tpl))
	if err != nil {
		log.Fatal(err)
	}

	var buffer bytes.Buffer
	err = t.Execute(&buffer, data)
	if err != nil {
		log.Fatal(err)
	}

	return buffer.String()
}
