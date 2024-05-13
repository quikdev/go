package context

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"time"

	fs "github.com/coreybutler/go-fsutil"
	"github.com/quikdev/go/command"
	"github.com/quikdev/go/config"
	"github.com/quikdev/go/util"
)

type Context struct {
	config              *config.Config
	Image               string            `json:"image,omitempty"`
	Env                 map[string]string `json:"environment_variables"`
	Variables           []string          `json:"ldflag_variables"`
	WASM                bool              `json:"wasm"`
	OS                  []string          `json:"operating_systems"`
	BuildFlags          []string          `json:"build_flags"`
	StripSymbols        bool              `json:"strip_symbols,omitempty"`
	StripDebugging      bool              `json:"disable_DWARF,omitempty"`
	LinkedLibPath       []string          `json:"linked_lib_path,omitempty"`
	ExtLD               string            `json:"extld,omitempty"`
	Shared              bool              `json:"shared"`
	ExtLDFlags          []string          `json:"extldflags,omitempty"`
	LDFlags             []string          `json:"ldflags,omitempty"`
	TempDir             string            `json:"tmp_dir,omitempty"`
	Tags                []string          `json:"tags,omitempty"`
	InstallDependencies bool              `json:"install_dependencies,omitempty"`
	Verbose             bool              `json:"verbose,omitempty"`
	CWD                 string            `json:"current_working_directory"`
	ASMFlags            *args             `json:"asm_flags"`
	GCCGoFlags          *args             `json:"gcc_go_flags"`
	GCFlags             *args             `json:"gc_flags"`
	Name                string            `json:"name"`
	OutputPath          string            `json:"output_dir"`
	OutputFileName      string            `json:"output_filename"`
	Tidy                bool              `json:"auto_tidy"`
	Port                int
	Tiny                bool `json:"use_tinygo"`
	UPX                 bool `json:"use_upx"`
	IgnoreCache         bool
	Cached              bool
	BuildFast           bool `json:"build_fast"`
}

func New() *Context {
	wd, err := os.Getwd()
	if err != nil {
		wd = "./"
	}

	return &Context{
		config:              config.New(),
		Env:                 make(map[string]string),
		Variables:           []string{},
		BuildFlags:          []string{},
		WASM:                false,
		OS:                  []string{runtime.GOOS},
		StripSymbols:        false,
		StripDebugging:      false,
		LinkedLibPath:       []string{},
		Shared:              false,
		ExtLDFlags:          []string{},
		Tags:                []string{},
		InstallDependencies: false,
		Verbose:             false,
		CWD:                 wd,
		OutputPath:          filepath.Join(wd, "./bin"),
		ASMFlags:            NewArgs("asmflags"),
		GCCGoFlags:          NewArgs("gccgoflags"),
		GCFlags:             NewArgs("gcflags"),
		Tidy:                false,
		Tiny:                false,
		UPX:                 false,
		IgnoreCache:         false,
		Cached:              false,
	}
}

func (ctx *Context) GetConfig() *config.Config {
	return ctx.config
}

func (ctx *Context) InputFile() string {
	file, err := findMainGoFile(ctx.CWD)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return file
}

func (ctx *Context) OutputFile() string {
	var name string
	if ctx.OutputFileName != util.EmptyString {
		name = ctx.OutputFileName
	} else {
		if val, exists := ctx.config.Get("build"); exists {
			name = val.(string)
		}

		if name != util.EmptyString {
			if obj, exists := ctx.config.Data()["variables"]; exists {
				if flagnm, exists := obj.(map[string]interface{})["build.name"]; exists {
					name = strings.ToLower(strings.ReplaceAll(flagnm.(string), " ", "_"))
				}

				if strings.HasPrefix(name, "package.") || strings.HasPrefix(name, "manifest.") {
					parts := strings.Split(name, ".")
					if flagnm, exists := ctx.config.Get(parts[len(parts)-1]); exists {
						name = strings.ToLower(strings.ReplaceAll(flagnm.(string), " ", "_"))
					}
				}
			}
		}

		if name == util.EmptyString {
			file, err := findMainGoFile(ctx.CWD)
			if err == nil {
				f := filepath.Base(file)
				name = f[:len(f)-len(filepath.Ext(f))]
			} else {
				name = filepath.Base(ctx.CWD)
			}
		}
	}

	name = name[:len(name)-len(filepath.Ext(name))]

	if runtime.GOOS == "windows" && !ctx.WASM {
		name += ".exe"
	}

	if ctx.WASM {
		name += ".wasm"
	}

	for strings.Contains(name, ".exe.exe") {
		name = strings.ReplaceAll(name, ".exe.exe", ".exe")
	}

	return name
}

func (ctx *Context) Output() string {
	if ctx.OutputPath == util.EmptyString {
		ctx.OutputPath = "./bin"
	}

	return filepath.Join(ctx.OutputPath, ctx.OutputFile())
}

func (ctx *Context) Configure() {
	// Configure Output File
	if out, exists := ctx.config.Get("name"); exists {
		regex := regexp.MustCompile("[^a-zA-Z0-9_-]+")
		ctx.OutputFileName = regex.ReplaceAllString(strings.ReplaceAll(out.(string), " ", "_"), "")
	}

	// Read linked variables from config
	if list, exists := ctx.config.Get("variables"); exists {
		for key, value := range list.(map[string]interface{}) {
			if strings.HasPrefix(strings.ToLower(value.(string)), "package.") || strings.HasPrefix(strings.ToLower(value.(string)), "manifest.") {
				subkey := strings.Split(value.(string), ".")[1:]
				if val, exists := ctx.config.Get(subkey[0]); exists {
					ctx.AddLinkedVariable(key, val.(string))
				} else {
					ctx.AddLinkedVariable(key, value.(string))
				}
			} else if strings.HasPrefix(strings.ToLower(value.(string)), "env.") {
				subkey := strings.Split(value.(string), ".")[1:]
				val := os.Getenv(subkey[0])
				ctx.AddLinkedVariable(key, val)
			} else {
				ctx.AddLinkedVariable(key, value.(string))
			}
		}
	}

	// Configured output
	// ctx.OutputPath = "./bin"
	if output, exists := ctx.config.Get("output"); exists {
		ctx.OutputPath = fs.Abs(output.(string))
	}
	if output, exists := ctx.config.Get("bin"); exists {
		ctx.OutputPath = fs.Abs(output.(string))
	}

	// Configured tags
	if tags, exists := ctx.config.Get("tags"); exists {
		for _, tag := range tags.([]interface{}) {
			ctx.AddTag(tag.(string))
		}
	}

	// Configured verbosity
	if verbose, exists := ctx.config.Get("verbose"); exists {
		ctx.Verbose = verbose.(bool)
	}

	// Configured autoinstall
	if install, exists := ctx.config.Get("install"); exists {
		ctx.InstallDependencies = install.(bool)
	}

	// Configure CWD
	if cwd, exists := ctx.config.Get("cwd"); exists {
		ctx.CWD = cwd.(string)
	}

	// Configure UPX (compress)
	if upx, exists := ctx.config.Get("compress"); exists {
		ctx.UPX = upx.(bool)
	}
	if upx, exists := ctx.config.Get("upx"); exists {
		ctx.UPX = upx.(bool)
	}

	// Configure other flags
	generics := []string{"a", "n", "race", "msan", "asan", "cover", "v", "work", "x", "modcacherw", "trimpath"}
	for _, item := range generics {
		name := strings.Replace(item, "-", "", 2)
		if _, exists := ctx.config.Get(name); exists {
			ctx.AddBuildFlag(item)
		}
	}

	simple := []string{"p", "covermode", "buildmode", "buildvcs", "compiler", "installsuffix", "mod", "modfile", "overlay", "pgo", "pkgdir", "toolexec"}
	for _, item := range simple {
		name := strings.Replace(item, "-", "", 2)
		if v, exists := ctx.config.Get(name); exists {
			ctx.AddBuildFlag(item, v.(string))
		}
	}

	splitFn := func(c rune) bool {
		return c == ' ' || c == '='
	}

	// Space-delimited args
	space := []string{"asmflags", "gccgoflags", "gcflags"}
	for _, flag := range space {
		if items, exists := ctx.config.Get(flag); exists {
			for _, item := range items.([]interface{}) {
				parts := strings.FieldsFunc(item.(string), splitFn)
				nm := parts[0]
				vals := []string{}
				if len(parts) > 1 {
					vals = parts[1:]
				}

				switch flag {
				case "asmflags":
					ctx.ASMFlags.Add(nm, vals...)
				case "gccgoflags":
					ctx.GCCGoFlags.Add(nm, vals...)
				case "gcflags":
					ctx.GCFlags.Add(nm, vals...)
				}
			}
			// for _, value := range items.([]string) {
			// 	for _, line := range strings.Split(value, " ") {
			// 		var rng []string
			// 		if strings.Contains(line, "=") {
			// 			rng = strings.Split(line, "=")
			// 		} else {
			// 			rng = strings.Split(line, " ")
			// 		}
			// 		for nm, val := range rng {
			// 			switch flag {
			// 			case "asmflags":
			// 				ctx.ASMFlags.Add(nm, val)
			// 			case "gccgoflags":
			// 				ctx.GCCGoFlags.Add(nm, val)
			// 			case "gcflags":
			// 				ctx.GCFlags.Add(nm, val)
			// 			}
			// 		}
			// 	}
			// }
			// for nm, val := range items.(map[string]string) {
			// 	vals := []string{}
			// 	if val != util.EmptyString {
			// 		vals = []string{val}
			// 	}

			// 	switch flag {
			// 	case "asmflags":
			// 		ctx.ASMFlags.Add(nm, vals...)
			// 	case "gccgoflags":
			// 		ctx.GCCGoFlags.Add(nm, vals...)
			// 	case "gcflags":
			// 		ctx.GCFlags.Add(nm, vals...)
			// 	}
			// }
		}
	}

	// Configure go mod tidy
	tidy, exists := ctx.config.Get("update")
	if exists {
		ctx.Tidy = tidy.(bool)
	} else {
		if tidy, exists = ctx.config.Get("tidy"); exists {
			ctx.Tidy = tidy.(bool)
		}
	}

	// Configure WASM builds
	if wasm, exists := ctx.config.Get("wasm"); exists {
		ctx.WASM = wasm.(bool)
	}

	// Configure build optimizations
	if buildfast, exists := ctx.config.Get("buildfast"); exists {
		ctx.BuildFast = buildfast.(bool)
	}

	// Configure port (WASM server)
	if port, exists := ctx.config.Get("port"); exists {
		ctx.Port = int(port.(float64))
	}

	// Configure tinygo
	if tgo, exists := ctx.config.Get("tiny"); exists {
		ctx.Tiny = tgo.(bool)
	}

	// Configure tinygo
	if minify, exists := ctx.config.Get("minify"); exists {
		val := minify.(bool)
		ctx.StripSymbols = val
		ctx.StripDebugging = val
	}

	if shrink, exists := ctx.config.Get("shrink"); exists {
		if shrink.(bool) {
			ctx.GCCGoFlags.Add("-s")
			ctx.GCCGoFlags.Add("-w")
		}
	}

	if nocache, exists := ctx.config.Get("no-cache"); exists {
		ctx.IgnoreCache = nocache.(bool)
	}

	if ldflags, exists := ctx.config.Get("ldflags"); exists {
		ldfs := ldflags.([]interface{})
		if len(ldfs) > 0 {
			ldf := []string{}
			for _, ldflag := range ldfs {
				ldf = append(ldf, ldflag.(string))
			}

			ctx.LDFlags = ldf
		}
	}
}

func (ctx *Context) isOutdated(bin string) bool {
	if !fs.Exists(bin) {
		return true
	}

	lastChange, err := util.GetLastChange(ctx.CWD)
	if err != nil {
		if strings.Contains(err.Error(), "cannot find the") {
			return true
		}

		util.Stderr(err)

		return false
	}

	binInfo, err := os.Stat(bin)
	if err != nil {
		util.Stderr(err)
		return false
	}

	return lastChange.After(binInfo.ModTime())
}

func (ctx *Context) BuildCommand(colorized ...bool) *command.Command {
	cmd := command.New()

	// Identify output file
	out := strings.Replace(ctx.Output(), ctx.CWD, ".", 1)
	if ctx.CWD == "./" {
		wd, _ := os.Getwd()
		out = strings.Replace(out, wd, "./", 1)
	}
	out = strings.ReplaceAll(out, "\\", "/")
	out = strings.Replace(out, ".go", "", 1)
	out = strings.Replace(strings.Replace(out, "./\\", ".\\", 1), ".//", "./", 1)

	if !ctx.isOutdated(out) && !ctx.IgnoreCache {
		ctx.Cached = true
		util.Stdout("# using cached build (qgo build --no-cache to force rebuild)\n")
		return cmd
	}

	if ctx.Tiny {
		cmd.Add("tinygo")
	} else {
		cmd.Add("go")
	}

	cmd.Add("build")

	if ctx.CWD != "./" {
		if ctx.CWD[0:1] == "\"" {
			ctx.CWD = ctx.CWD[1:]
		}
		if ctx.CWD[:len(ctx.CWD)-1] == "\"" {
			ctx.CWD = ctx.CWD[:len(ctx.CWD)-2]
		}
		cwd, err := os.Getwd()
		if err == nil {
			ctx.CWD = strings.Replace(ctx.CWD, cwd, "", 1)
			if ctx.CWD == "" {
				ctx.CWD = "./"
			}
		}

		if ctx.CWD != "./" {
			cmd.Add("-C", "\""+ctx.CWD+"\"")
		}
	}

	if ctx.WASM {
		os.Setenv("GOOS", "js")
		os.Setenv("GOARCH", "wasm")
	}

	// Linked Flags
	ctx.AddLinkedVariable("main.buildTime", time.Now().UTC().Format(time.RFC3339))

	if ctx.StripDebugging {
		ctx.AddLinkedFlag("-w")
	}

	if ctx.StripSymbols {
		ctx.AddLinkedFlag("-s")
	}

	for _, lib := range ctx.LinkedLibPath {
		ctx.AddLinkedFlag("-L" + lib)
	}

	if ctx.Shared {
		ctx.AddLinkedFlag("-shared")
	}

	if len(ctx.ExtLDFlags) > 0 {
		ctx.AddLinkedFlag("-linkshared \"" + strings.Join(ctx.ExtLDFlags, " ") + "\"")
	}

	if ctx.TempDir != util.EmptyString {
		ctx.AddLinkedFlag("-tmpdir " + ctx.TempDir)
	}

	ldflags := ctx.LDFlags
	if ctx.LDFlags == nil {
		ldflags = []string{}
	}
	if len(ctx.Variables) > 0 {
		for _, flag := range ctx.Variables {
			ldflags = append(ldflags, flag)
		}
	}

	// Concatenate command
	if ctx.InstallDependencies {
		cmd.Add("-i")
	}

	if ctx.Verbose {
		cmd.Add("-v")
	}

	if len(ldflags) > 0 {
		cmd.Add("-ldflags", "\""+strings.Join(ldflags, " ")+"\"")
	}

	// Add tags
	if len(ctx.Tags) > 0 {
		cmd.Add("-tags=" + strings.Join(ctx.Tags, ","))
		// cmd.Add("-tags \"" + strings.Join(ctx.Tags, " ") + "\"")
	}

	// Add space delimited args
	if ctx.ASMFlags.Size() > 0 {
		cmd.Add(ctx.ASMFlags.String())
	}

	if ctx.GCCGoFlags.Size() > 0 {
		cmd.Add(ctx.GCCGoFlags.String())
	}

	if ctx.GCFlags.Size() > 0 {
		cmd.Add(ctx.GCFlags.String())
	}

	// Add output
	cmd.Add("-o", out)

	if ctx.Tiny {
		cmd.Add("-target=wasm")
	}

	cmd.Add(strings.TrimSpace(strings.ReplaceAll(ctx.InputFile(), " ", "\\ ")))

	return cmd
}

func (ctx *Context) RunCommand(colorized ...bool) *command.Command {
	cmd := ctx.BuildCommand(colorized...)

	if !ctx.WASM {
		out := strings.Replace(ctx.Output(), ctx.CWD, ".", 1)
		out = strings.Replace(out, ".go", "", 1)
		// out = strings.ReplaceAll(out, "\\", "/")
		if len(strings.TrimSpace(cmd.String())) > 0 {
			cmd.Add("&&")
		}
		cmd.Add(out)

		// append arguments supplied by the user
		// ignore the build file, or the `--` separator if it exists.
		args := os.Args[2:]
		if len(args) > 0 {
			index := util.IndexOf[string](args, "--")
			if index >= 0 {
				args = args[index+1:]
			} else if len(args) > 0 {
				if filepath.Ext(args[0]) == ".go" {
					args = args[1:]
				}
			}
		}

		// Ignore the no-cache flag
		ignoreList := []string{"bundle", "os", "wasm", "output", "tips", "minify", "shrink", "dry-run", "nowork", "update", "port", "no-cache"}
		for _, ignored := range ignoreList {
			i := util.IndexOf[string](args, "--"+ignored)
			if i >= 0 {
				args = slices.Delete(args, i, i+1)
			}
		}

		if len(args) > 0 {
			cmd.Add(args...)
		}
	}

	return cmd
}

func (ctx *Context) AddBuildFlag(name string, value ...string) {
	val := name
	if len(value) > 0 {
		val += " " + value[0]
	}

	if !util.InSlice[string](val, ctx.BuildFlags) {
		ctx.BuildFlags = append(ctx.BuildFlags, val)
	}
}

func (ctx *Context) AddLinkedVariable(key string, value string) {
	val := "-X '" + key + "=" + value + "'"
	if !util.InSlice[string](val, ctx.Variables) {
		ctx.Variables = append(ctx.Variables, val)
	}
}

func (ctx *Context) AddLinkedFlag(value string) {
	if !util.InSlice[string](value, ctx.Variables) {
		ctx.Variables = append(ctx.Variables, value)
	}
}

func (ctx *Context) AddTag(tag string) {
	if !util.InSlice[string](tag, ctx.Tags) {
		ctx.Tags = append(ctx.Tags, tag)
	}
}

func findMainGoFile(dir string) (string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		filePath := filepath.Join(dir, file.Name())

		if file.IsDir() {
			// Recursively search in subdirectories.
			mainFile, err := findMainGoFile(filePath)
			if err == nil && mainFile != "" {
				return mainFile, nil
			}
		} else if strings.HasSuffix(file.Name(), ".go") {
			// Check if the .go file contains "func main()".
			if containsMainFunction(filePath) {
				return filePath, nil
			}
		}
	}

	return "", nil
}

func containsMainFunction(filePath string) bool {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return false
	}

	return strings.Contains(string(data), "func main()")
}
