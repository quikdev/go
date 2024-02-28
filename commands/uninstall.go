package commands

import (
	"debug/pe"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/huh"
	fs "github.com/coreybutler/go-fsutil"
	"github.com/quikdev/go/v1/util"
)

type Uninstall struct {
	NoWarn  bool   `name:"no-warn" type:"bool" help:"Will not warn/prompt for approval before removing the app."`
	AppName string `arg:"name" optional:"" help:"Name of the app to uninstall."`
}

func (u *Uninstall) Run(c *Context) error {
	cmd := exec.Command("go", "env")
	out, err := cmd.Output()
	util.BailOnError(err)

	lines := strings.Split(string(out), "\n")

	// Initialize variables to store values
	var GOBIN, GOPATH string

	// Iterate through the lines and extract the values
	for _, line := range lines {
		if strings.HasPrefix(line, "set GOBIN=") {
			GOBIN = strings.TrimSpace(strings.TrimPrefix(line, "set GOBIN="))
		} else if strings.HasPrefix(line, "set GOPATH=") {
			GOPATH = strings.TrimSpace(strings.TrimPrefix(line, "set GOPATH="))
			if len(GOPATH) > 0 {
				GOPATH = filepath.Join(GOPATH, "bin")
			}
		}
	}

	if len(GOBIN) == 0 && len(GOPATH) == 0 {
		util.Stderr("cannot find a global installation root in GOBIN, GOPATH, or GOROOT", true)
	}

	roots := splitFilePaths(GOBIN)
	roots = append(roots, splitFilePaths(GOPATH)...)

	fulllist := []string{}
	for _, dir := range roots {
		if fs.Exists(dir) {
			options, err := findExecutableFiles(dir)
			util.BailOnError(err)

			if len(u.AppName) > 0 {
				for _, opt := range options {
					if getFilenameWithoutExtension(opt) == u.AppName {
						u.uninstall(opt)
					}
				}
			}

			fulllist = append(fulllist, options...)
		}
	}

	if len(u.AppName) == 0 {
		opts := make([]huh.Option[string], len(fulllist))
		for i, opt := range fulllist {
			opts[i] = huh.NewOption[string](getFilenameWithoutExtension(opt), opt)
		}

		util.Highlight("No app specified.")
		var selected string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Which Go application should be removed?").
					Options(opts...).
					Value(&selected),
			),
		)

		util.BailOnError(form.Run())

		u.uninstall(selected)
	} else {
		util.Stderr("not found - try running \"qgo uninstall\" (without \""+u.AppName+"\") to select from a list of known apps", true)
	}

	return nil
}

func (u *Uninstall) uninstall(path string) {
	if !fs.Exists(path) {
		util.Stderr(path+" does not exist", true)
	}

	rm := false
	if !u.NoWarn {
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Are you sure you want to uninstall " + path + "?").
					Affirmative("Yes, Remove").
					Negative("No").
					Value(&rm),
			),
		)

		util.BailOnError((form.Run()))
	} else {
		rm = true
	}

	if rm {
		err := os.Remove(path)
		util.BailOnError(err)
		util.Stdout(fmt.Sprintf("# removed %v\n", path))
	}

	os.Exit(0)
}

func findExecutableFiles(dirPath string) ([]string, error) {
	var executableFiles []string

	// Open the directory
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	// Read the directory entries
	files, err := dir.Readdir(0)
	if err != nil {
		return nil, err
	}

	// Iterate over the files
	for _, file := range files {
		// Check if the file is executable
		abspath := filepath.Join(dirPath, file.Name())
		switch runtime.GOOS {
		case "windows":
			if isExecutableWindows(abspath) {
				executableFiles = append(executableFiles, abspath)
			}
		case "darwin", "linux":
			if file.Mode()&0111 != 0 {
				executableFiles = append(executableFiles, abspath)
			}
		}
	}

	return executableFiles, nil
}

func isExecutableWindows(filename string) bool {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer file.Close()

	// Attempt to parse the PE file
	peFile, err := pe.NewFile(file)
	if err != nil {
		return false
	}

	// Verify the PE signature
	if peFile.Machine == pe.IMAGE_FILE_MACHINE_UNKNOWN {
		return false
	}

	// Check for "MZ" header
	mzHeader := make([]byte, 2)
	_, err = file.ReadAt(mzHeader, 0)
	if err != nil {
		return false
	}
	if string(mzHeader) != "MZ" {
		return false
	}

	// If all checks pass, return true
	return true
}

func splitFilePaths(input string) []string {
	if runtime.GOOS == "windows" {
		// Split the input string by colons
		paths := strings.Split(input, ":")

		// Reconstruct the file paths
		var filepaths []string
		for _, path := range paths {
			// Check if the path starts with a drive letter
			if len(path) > 2 && path[1] == '\\' && path[2] == '\\' {
				// If it does, append it directly
				filepaths = append(filepaths, path)
			} else if len(filepaths) > 0 {
				// If not, append to the last entry in the file paths slice
				filepaths[len(filepaths)-1] += ":" + path
			} else {
				// If there's no previous entry, just append
				filepaths = append(filepaths, path)
			}
		}

		return filepaths
	}

	return strings.Split(input, ":")
}

func getFilenameWithoutExtension(filePath string) string {
	// Use the filepath.Base function to get the base (filename) from the absolute path
	filename := filepath.Base(filePath)

	// Use the filepath.Ext function to get the extension of the filename
	extension := filepath.Ext(filename)

	// Remove the extension from the filename
	filenameWithoutExt := filename[:len(filename)-len(extension)]

	return filenameWithoutExt
}
