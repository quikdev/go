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
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/quikdev/go/context"
	"github.com/quikdev/go/util"

	fs "github.com/coreybutler/go-fsutil"
)

type Run struct {
	Bundle      []string `name:"bundle" short:"b" type:"string" help:"Bundle the application into a tarball/zipball" enum:"zip,tar"`
	OS          []string `name:"os" type:"string" help:"The operating system(s) to build for (any option(s) from 'go tool dist list' is valid)"`
	WASM        bool     `name:"wasm" type:"bool" help:"Output a web assembly (OS is ignored when this option is true)"`
	Output      string   `name:"output" short:"o" type:"string" help:"Output file name"`
	Tips        bool     `name:"tips" short:"t" type:"bool" help:"Display tips in the generated commands"`
	Minify      bool     `name:"minify" short:"m" type:"bool" help:"Set ldflags to strip debugging symbols and remove DWARF generations"`
	Shrink      bool     `name:"shrink" short:"s" type:"bool" help:"Set gccgoflags to strip debugging symbols and remove DWARF generations"`
	DryRun      bool     `name:"dry-run" short:"d" type:"bool" help:"Display the command without executing it."`
	NoWork      bool     `name:"no-work" type:"bool" help:"Set GOWORK=off when building"`
	Update      bool     `name:"update" short:"u" type:"bool" help:"Update (go mod tidy) before building."`
	Port        int      `name:"port" short:"p" help:"The port to run the HTTP server on (WASM only)."`
	IgnoreCache bool     `name:"no-cache" type:"bool" help:"Ignore the cache and rebuild, even if no Go files have changed."`
	Profile     []string `name:"profile" optional:"" help:"Name of the manifest.json profile attribute to apply."`
	File        string   `arg:"source" optional:"" help:"Go source file (ex: main.go)"`
	Args        []string `arg:"" optional:"" help:"Arguments to pass to the executable."`
	// Container string `name:"container" default:"docker" type:"string" enum:"docker,podman" help:"The containerization technology to build with"`
}

func (b *Run) Run(c *Context) error {
	ctx := context.New(b.Profile...)
	ctx.Configure()

	if len(strings.TrimSpace(ctx.InputFile())) == 0 {
		_, err := util.FindMainFileInDirectory("./")
		if err != nil {
			util.Stderr(err)
			util.SubtleHighlight("Only apps with a 'package main' can be run (modules cannot be run directly).")
			os.Exit(1)
		}
	}

	if b.IgnoreCache {
		ctx.IgnoreCache = b.IgnoreCache
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

	// Run this before go mod tidy to determine whether the output is cached
	cmd := ctx.RunCommand()
	if c.Debug {
		b.Tips = true
	}

	// Display commands
	hide := false
	hideDisplay, err := c.Get("hide")
	if err == nil {
		if hideDisplay != nil {
			hide = hideDisplay.(bool)
		}
	}

	if !b.DryRun {
		if ctx.Tidy {
			if !ctx.Cached {
				if hide {
					util.BailOnError(util.StreamNoStdErrNoHighlight("go mod tidy"))
				} else {
					util.BailOnError(util.StreamNoStdErr("go mod tidy"))
					fmt.Println("")
				}
			}
		}

		vars := ctx.GetConfig().GetEnvVars()
		if len(vars) > 0 {
			util.Stdout(`# autoapplying the following environment variables` + "\n")
			for key, value := range vars {
				os.Setenv(key, value)
				util.Stdout("\n  " + strings.Replace(util.SubtleHighlighter(key), "\n", "", 1) + util.Dim("=") + strings.ReplaceAll(util.Highlighter(value), "\n", ""))
			}
			fmt.Printf("\n\n")
		}
	}

	if !hide {
		fmt.Println(cmd.Display(b.Tips) + "\n")
	}

	parentdir := filepath.Dir(ctx.Output())
	if !fs.Exists(parentdir) {
		os.MkdirAll(parentdir, os.ModePerm)
	}

	// Run command
	if !b.DryRun {
		if b.NoWork {
			os.Setenv("GOWORK", "off")
		}

		subscribers := make(map[chan string]bool)
		reload, livereloadexists := ctx.GetConfig().Get("livereload")
		if !livereloadexists {
			tmp := make([]interface{}, 2)
			var iface interface{}
			iface = "**/*.go"
			tmp[0] = iface
			iface = "*.go"
			tmp[1] = iface
			iface = tmp
			reload = iface
		}

		// Placeholder to prevent duplicate event handling
		ignoreEvents := false

		for _, glob := range reload.([]interface{}) {
			matches, err := filepath.Glob(glob.(string))
			if err != nil {
				util.Stderr(err)
			} else {
				for _, match := range matches {
					watcher, err := fsnotify.NewWatcher()
					if err != nil {
						util.Stderr(err)
					} else {
						defer watcher.Close()

						go func() {
							for {
								select {
								case event, ok := <-watcher.Events:
									if !ok {
										return
									}

									if event.Has(fsnotify.Write) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) || event.Has(fsnotify.Create) {
										if ctx.WASM {
											if !ignoreEvents && event.Has(fsnotify.Write) {
												go func() {
													ignoreEvents = true
													ctx.IgnoreCache = true
													cmd = ctx.BuildCommand()
													fmt.Println(cmd.Display())
													cmd.Run(ctx.CWD)
													// fmt.Println("execution complete")
													time.Sleep(500 * time.Millisecond)
													for subscriber := range subscribers {
														subscriber <- "reload"
													}
													time.Sleep(3 * time.Second)
													ignoreEvents = false
												}()
											}
										} else {
											watcher.Close()

											// c.Set("hide", true)
											util.SubtleHighlight("rebuilding/running " + ctx.OutputFile())
											// fmt.Printf("kill %v\n", cmd.PID())
											b.Run(c)
										}
									}
								case err, ok := <-watcher.Errors:
									if !ok {
										return
									}
									util.Stderr(err, true)
								}
							}
						}()

						err := watcher.Add(match)
						if err != nil {
							util.Stderr(err)
						}
					}
				}
			}
		}

		if ctx.WASM {
			cmd.Run(ctx.CWD)

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

			http.HandleFunc("/livereload", func(w http.ResponseWriter, r *http.Request) {
				util.SubtleHighlight(r.Header.Get("User-Agent") + " connected")
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("Cache-Control", "no-cache")
				w.Header().Set("Connection", "keep-alive")

				clientChan := make(chan string)
				subscribers[clientChan] = true
				defer func() { delete(subscribers, clientChan) }()
				defer r.Body.Close()

				for {
					select {
					case event := <-clientChan:
						// fmt.Println(`sent: ` + event)
						w.Write([]byte("data: " + event + "\n\n"))
						w.(http.Flusher).Flush()
					case <-r.Context().Done():
						delete(subscribers, clientChan)
						util.Stdout("\n" + r.UserAgent() + " disconnected")
						return
					}
				}
			})

			http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/" {
					http.ServeFile(w, r, filepath.Join(root, "index.html"))
					return
				}

				server.ServeHTTP(w, r)
			})

			url := fmt.Sprintf("http://localhost:%d", port)

			// Launch test server
			go func() {
				defer wg.Done()
				fmt.Println("")
				util.HighlightComment("launching QuikGo HTTP server...")
				util.Stdout(fmt.Sprintf("server available at %s\nctrl+c or cmd+c to quit\n", url))
				util.BailOnError(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
			}()

			// Optionally open browser
			autobrowse := true
			if value, exists := ctx.GetConfig().Get("autobrowse"); exists {
				autobrowse = value.(bool)
			}

			if autobrowse {
				go openbrowser(url)
			}

			wg.Wait()
		} else {
			// fmt.Println(cmd.String())
			cmd.Run(ctx.CWD)

			// Forcibly exit the process when the command finishes
			// (so the reload mechanism will close)
			os.Exit(0)
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

func openbrowser(url string) {
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
}
