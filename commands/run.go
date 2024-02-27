package commands

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/quikdev/go/v1/context"
	"github.com/quikdev/go/v1/util"

	fs "github.com/coreybutler/go-fsutil"
)

type Run struct {
	Bundle []string `name:"bundle" short:"b" type:"string" help:"Bundle the application into a tarball/zipball" enum:"zip,tar"`
	OS     []string `name:"os" type:"string" help:"The operating system(s) to build for (any options from 'go tool dist list' is valid)"`
	WASM   bool     `name:"wasm" type:"bool" help:"Output a web assembly (OS is ignored when this option is true)"`
	Output string   `name:"output" short:"o" type:"string" help:"Output file name"`
	Tips   bool     `name:"tips" short:"t" type:"bool" help:"Display tips in the generated commands"`
	Minify bool     `name:"minify" short:"m" type:"bool" help:"Set ldflags to strip debugging symbols and remove DWARF generations"`
	Shrink bool     `name:"shrink" short:"s" type:"bool" help:"Set gccgoflags to strip debugging symbols and remove DWARF generations"`
	DryRun bool     `name:"dry-run" short:"d" type:"bool" help:"Display the command without executing it."`
	NoWork bool     `name:"nowork" type:"bool" help:"Set GOWORK=off when building"`
	Update bool     `name:"update" short:"u" type:"bool" help:"Update (go mod tidy) before building."`
	Port   int      `name:"port" short:"p" help:"The port to run the HTTP server on (WASM only)."`
	File   string   `arg:"source" optional:"" help:"Go source file (ex: main.go)"`
	// Container string `name:"container" default:"docker" type:"string" enum:"docker,podman" help:"The containerization technology to build with"`
}

func (b *Run) Run(c *Context) error {
	ctx := context.New()
	ctx.Configure()

	if len(strings.TrimSpace(ctx.InputFile())) == 0 {
		_, err := util.FindMainFileInDirectory("./")
		if err != nil {
			util.Stderr(err)
			util.SubtleHighlight("Only apps with a 'package main' can be run (modules cannot be run directly).")
			os.Exit(1)
		}
	}

	if b.Update {
		ctx.Tidy = true
	}

	if b.WASM {
		ctx.OS = []string{"js"}
	} else {
		ctx.OS = b.OS
	}

	if b.Minify {
		ctx.StripSymbols = true
		ctx.StripDebugging = true
	}

	if b.Shrink {
		ctx.GCCGoFlags.Add("-s")
		ctx.GCCGoFlags.Add("-w")
	}

	if !b.DryRun {
		if ctx.Tidy {
			util.BailOnError(util.Run("go mod tidy"))
			fmt.Println("")
		}
	}

	cmd := ctx.RunCommand()
	if c.Debug {
		b.Tips = true
	}

	// Display command
	fmt.Println(cmd.Display(b.Tips))

	parentdir := filepath.Dir(ctx.Output())
	if !fs.Exists(parentdir) {
		os.MkdirAll(parentdir, os.ModePerm)
	}

	// Run command
	if !b.DryRun {
		if b.NoWork {
			os.Setenv("GOWORK", "off")
		}

		if ctx.WASM {
			root := filepath.Dir(ctx.Output())

			port := b.Port
			if port == util.EmptyInt {
				port = ctx.Port
				if port == util.EmptyInt {
					var err error
					port, err = findOpenPort(8000, 65535)
					util.BailOnError(err)
				}
			}

			var wg sync.WaitGroup

			wg.Add(1)

			server := http.FileServer(http.Dir(root))

			http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/" {
					http.ServeFile(w, r, filepath.Join(root, "index.html"))
					return
				}

				server.ServeHTTP(w, r)
			})

			url := fmt.Sprintf("http://localhost:%d", port)

			go func() {
				defer wg.Done()
				fmt.Println("")
				util.HighlightComment("launching QuikGo HTTP server...")
				util.Stdout(fmt.Sprintf("server available at %s\nctrl+c or cmd+c to quit\n", url))
				util.BailOnError(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
			}()

			var cmd string
			switch runtime.GOOS {
			case "windows":
				cmd = "cmd"
				args := []string{"/c", "start", url}
				exec.Command(cmd, args...).Start()
			case "darwin":
				cmd = "open"
				exec.Command(cmd, url).Start()
			default:
				cmd = "xdg-open"
				exec.Command(cmd, url).Start()
			}

			wg.Wait()
		} else {
			cmd.Run(ctx.CWD)
		}
	}

	return nil
}

func findOpenPort(start, end int) (int, error) {
	for port := start; port <= end; port++ {
		address := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", address)
		if err == nil {
			listener.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no open port found in range %d-%d", start, end)
}
