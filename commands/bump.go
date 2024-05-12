package commands

import (
	"fmt"
	"os"
	"regexp"

	"github.com/Masterminds/semver"
	"github.com/quikdev/go/context"
	"github.com/quikdev/go/util"
)

type Bump struct {
	Type string `arg:"type" optional:"" default:"patch" enum:"major,minor,patch,prerelease,pre,build" help:"major, minor, patch (default), prerelease, or build"`
}

// TODO: test this
// with multiline support
func (b *Bump) Run(c *Context) error {
	ctx := context.New()
	ctx.Configure()
	cfg := ctx.GetConfig()

	// Get the manifest version
	rawversion, exists := cfg.Get("version")
	version := "0.0.0"
	if exists {
		version = rawversion.(string)
	}

	v, err := semver.NewVersion(version)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var newVersion semver.Version

	switch b.Type {
	case "major":
		newVersion = v.IncMajor()
	case "minor":
		newVersion = v.IncMinor()
	case "patch":
		newVersion = v.IncPatch()
	case "pre":
		fallthrough
	case "prerelease":
		fmt.Println("prelrelease version bumping not currently supported")
		os.Exit(0)
	case "build":
		fmt.Println("build version bumping not currently supported")
		os.Exit(0)
	}

	original := rawversion
	if rawversion == nil {
		var s interface{}
		s = "none"
		original = s
	}

	// Read in file
	input, err := cfg.Raw()
	if err != nil {
		util.Stderr(err, true)
	}

	regex := regexp.MustCompile(`"version"\s*:\s*"([^"]+)"`)
	output := regex.ReplaceAllString(string(input), `"version": "`+newVersion.String()+`"`)

	os.WriteFile(cfg.File(), []byte(output), os.ModePerm)

	fmt.Printf("bumped %s → %s\n", original.(string), util.Highlighter(newVersion.String()))

	return nil
}
