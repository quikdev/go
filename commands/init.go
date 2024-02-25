package commands

import (
	"archive/zip"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/quikdev/qgo/v1/context"
	"github.com/quikdev/qgo/v1/util"
)

type Init struct {
	Overwrite bool `name:"overwrite" short:"o" help:"If a project already exist, overwrite with no backup and no prompt."`
}

var (
	appname     string
	modulename  string
	description string
	license     string
	githubuser  string
	location    string
)

//go:embed assets
var assetsFS embed.FS

func (i *Init) Run(c *Context) error {
	ctx := context.New()
	ctx.Configure()

	// COLLECT DATA
	var err error
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "./"
	}

	location = cwd

	// Create a new HTTP client with a timeout
	client := &http.Client{
		Timeout: 1500 * time.Millisecond,
	}

	var licenses []huh.Option[string]
	var licensedata []map[string]string

	// Make GET request to the GitHub API endpoint using the custom client
	resp, err := client.Get("https://api.github.com/licenses")
	if err == nil {
		// Decode the JSON response
		if err := json.NewDecoder(resp.Body).Decode(&licensedata); err == nil {
			// Populate the licenses
			for _, l := range licensedata {
				licenses = append(licenses, huh.NewOption[string](l["spdx_id"]+": "+l["name"], l["spdx_id"]))
			}
		}

		resp.Body.Close()
	}

	licenses = append(licenses, huh.NewOption[string]("Custom", "custom"))
	licenses = append(licenses, huh.NewOption[string]("No License", "none"))

	createtype := "command"

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project Name (ex: myapp)").
				// Prompt("? ")
				Validate(func(val string) error {
					if len(strings.TrimSpace(val)) == 0 {
						return errors.New("Name cannot be blank.")
					}

					if strings.HasSuffix(strings.ToLower(val), ".exe") {
						return errors.New("Do not add the .exe extension (this is added automatically to Windows executables)")
					}

					return nil
				}).
				Value(&appname),
			huh.NewText().
				Title("Description").
				Value(&description),
			huh.NewSelect[string]().
				Title("What are you creating?").
				Options(
					huh.NewOption("Go Module", "module"),
					huh.NewOption("App/CLI", "command"),
					huh.NewOption("Web Assembly (WASM)", "wasm"),
					// huh.NewOption("Basic package + internal packages", "module_supported"),
					// huh.NewOption("Basic command", "command"),
					// huh.NewOption("Basic command + internal packages", "command_supported"),
					// huh.NewOption("Package(s) + Command(s)", "module_command"),
					// huh.NewOption("Package(s) + Command(s) + internal packages", "module_command_supported"),
					// huh.NewOption("Multiple packages", "multipackage"),
					// huh.NewOption("Multiple commands", "multicommand"),
					// huh.NewOption("Server (ex HTTP/API)", "server_command"),
				).
				Value(&createtype),
		),
		huh.NewGroup(
			huh.NewSelect[string]().Options(licenses...).Title("Pick a license.").Value(&license),
		),
	)
	util.BailOnError(form.Run())

	// Get default module name
	_, email, err := util.GetGitHubInfo()
	if err == nil {
		githubuser = util.ExtractGitHubHandleFromEmail(email)
		if len(githubuser) > 0 {
			if err == nil {
				modulename = "github.com/" + githubuser + "/" + appname //filepath.Base(cwd)
			}
		}
	}

	form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Module Name").
				// Prompt("? ").
				Validate(func(val string) error {
					if len(strings.TrimSpace(val)) == 0 {
						return errors.New("Module name cannot be blank.")
					}

					return nil
				}).
				Value(&modulename),
		),
	)
	util.BailOnError(form.Run())

	location = strings.Replace(filepath.Join(location, appname), cwd+string(filepath.Separator), "."+string(filepath.Separator), 1)
	usegit := true

	form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Where should the new project be created?").
				Value(&location),
			huh.NewConfirm().
				Title("Configure git?").
				Affirmative("Yes").
				Negative("No").
				Value(&usegit),
		),
	)
	util.BailOnError(form.Run())

	// Setup git repo
	remoteorigin := false
	remoterepo := ""
	if usegit {
		form = huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Add remote git repo?").
					Affirmative("Yes").
					Negative("No").
					Value(&remoteorigin),
			),
		)
		form.Run()

		if remoteorigin {
			if strings.Contains(modulename, "github.com") {
				remoterepo = modulename
			}

			form = huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Provide the remote origin repo:").
						Validate(func(val string) error {
							if len(strings.TrimSpace(val)) == 0 {
								return errors.New("remote origin cannot be blank - expected a value like git@github.com:user/repo.git")
							}
							if !strings.HasSuffix(val, ".git") {
								return errors.New("expected a value like git@domain.com:user/repo.git")
							}

							return nil
						}).
						Value(&remoterepo),
				),
			)
			form.Run()
		}
	}

	// INITIALIZE PROJECT
	// Create project directory
	util.HighlightCommand("mkdir -p", location)
	err = os.MkdirAll(location, os.ModePerm)
	if err != nil {
		util.Stderr(err)
		return nil
	}

	util.HighlightCommand("cd", location)

	abspath, err := filepath.Abs(location)
	util.BailOnError(err)

	// Check for existing go module
	if util.FileExists(filepath.Join(abspath, "go.mod")) || util.FileExists(filepath.Join(abspath, "package.json")) || util.FileExists(filepath.Join(abspath, "manifest.json")) {
		recreate := false
		backup := true
		if i.Overwrite == true {
			recreate = true
			backup = false
		} else {
			form = huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title("A module and/or manifest already exists. Do you want to backup & recreate the core project?").
						Affirmative("Yes").
						Negative("No").
						Value(&recreate),
				),
			)
			util.BailOnError(form.Run())
		}

		if !recreate {
			util.Stdout("aborted to prevent conflict", true)
		}

		// Archive the old module
		list := []string{"go.mod", "go.sum", "go.work", "package.json", "manifest.json", "README.md", "LICENSE", ".gitignore", appname + ".go", "main.go", "internal", "cmd", "submodule", "module_a", "module_b", "greeting"}
		if backup {
			files := []string{}
			for _, item := range list {
				p := filepath.Join(abspath, item)
				if util.FileExists(p) {
					files = append(files, p)
				}
			}

			ms := time.Now().UnixNano() / int64(time.Millisecond)
			filename := fmt.Sprintf("%s-backup-%v.zip", appname, ms)
			zipFile, err := os.Create(filepath.Join(abspath, filename))
			util.BailOnError(err)
			defer zipFile.Close()

			zipWriter := zip.NewWriter(zipFile)
			defer zipWriter.Close()

			for _, file := range files {
				err := addFileToZip(zipWriter, abspath, file)
				util.BailOnError(err)
			}
		}

		// Remove conflicting files after backup
		for _, item := range list {
			p := filepath.Join(abspath, item)
			if util.FileExists(p) {
				err := os.RemoveAll(p)
				if err != nil {
					util.Stderr(err)
				} else {
					util.HighlightCommand("rm", p)
				}
			}
		}
	}

	// generate go.mod file
	util.BailOnError(util.Run("go mod init "+modulename, abspath))

	author := githubuser
	if len(githubuser) == 0 {
		currentUser, err := user.Current()
		if err == nil {
			author = currentUser.Username
		}
	}

	// generate README
	source, _ := assetsFS.ReadFile("assets/readme.tpl")
	content := util.ApplyTemplate(source, map[string]string{
		"Name":        appname,
		"Description": description,
		// "Module":      modulename,
		"Year":   fmt.Sprintf("%d", time.Now().Year()),
		"Author": author,
	})
	util.WriteTextFile(filepath.Join(abspath, "README.md"), content, true)

	// generate the LICENSE
	if license == "custom" {
		util.WriteTextFile(filepath.Join(abspath, "LICENSE"), "TODO: Add custom license", true)
		util.Highlight("Remember to add your custom license to the LICENSE file.")
	} else if license == "none" {
		util.Highlight("WARNING: No license file generated.")
	} else {
		for _, lic := range licensedata {
			if lic["spdx_id"] == license {
				resp, err := http.Get(lic["url"])
				if err == nil {
					defer resp.Body.Close()

					var data map[string]interface{}
					decoder := json.NewDecoder(resp.Body)
					if err := decoder.Decode(&data); err != nil && err != io.EOF {
						util.Stderr("failed to create LICENSE file: " + err.Error())
						break
					}

					body := data["body"].(string)
					year := fmt.Sprintf("%d", time.Now().Year())

					body = strings.ReplaceAll(body, "<year>", year)
					body = strings.ReplaceAll(body, "[year]", year)
					body = strings.ReplaceAll(body, "[yyyy]", year)
					body = strings.ReplaceAll(body, "<program>", appname)
					body = strings.ReplaceAll(body, "<name of author>", author)
					body = strings.ReplaceAll(body, "[name of author]", author)
					body = strings.ReplaceAll(body, "[name of copyright owner]", author)
					body = strings.ReplaceAll(body, "[fullname]", author)
					body = strings.ReplaceAll(body, "[author]", author)

					if len(description) == 0 {
						body = strings.ReplaceAll(body, "<one line to give the program's name and a brief idea of what it does.>", appname+" application")
					} else {
						body = strings.ReplaceAll(body, "<one line to give the program's name and a brief idea of what it does.>", description)
					}

					util.WriteTextFile(filepath.Join(abspath, "LICENSE"), body, true)
					util.SubtleHighlight("      ↳ review the LICENSE to assure your name and program details are defined correctly.")
				} else {
					util.Stderr("failed to create LICENSE file: " + err.Error())
				}
				break
			}
		}
	}

	mod := strings.Split(modulename, "/")[len(strings.Split(modulename, "/"))-1]
	if createtype == "module" {
		// generate module file
		source, _ = assetsFS.ReadFile("assets/module.go.tpl")
		content = util.ApplyTemplate(source, map[string]string{
			"Module": mod,
		})
		util.WriteTextFile(filepath.Join(abspath, appname+".go"), content, true)

		// generate module test file
		source, _ = assetsFS.ReadFile("assets/module_test.go.tpl")
		content = util.ApplyTemplate(source, map[string]string{
			"Module":      mod,
			"Name":        appname,
			"Version":     "0.0.1",
			"Description": description,
		})
		util.WriteTextFile(filepath.Join(abspath, appname+"_test.go"), content, true)
	} else if createtype == "wasm" {
		// generate main file
		source, _ = assetsFS.ReadFile("assets/wasm.go.tpl")
		content = util.ApplyTemplate(source, map[string]string{})
		util.WriteTextFile(filepath.Join(abspath, "main.go"), content, true)

		// generate HTML file
		source, _ = assetsFS.ReadFile("assets/wasm.html.tpl")
		content = util.ApplyTemplate(source, map[string]string{
			"Name": mod,
		})
		os.MkdirAll(filepath.Join(abspath, "bin"), os.ModePerm)
		util.WriteTextFile(filepath.Join(abspath, "bin", "index.html"), content, true)

		// Get the wasm_exec.js file
		url := "https://raw.githubusercontent.com/golang/go/master/misc/wasm/wasm_exec.js"
		resp, err := http.Get(url)
		msg := "The wasm_exec.js file is available at " + url
		if err != nil {
			util.Stderr(err)
			util.SubtleHighlight(msg)
		} else {
			defer resp.Body.Close()
			file, err := os.Create(filepath.Join(abspath, "bin", "wasm_exec.js"))
			if err != nil {
				util.Stderr(err)
				util.SubtleHighlight(msg)
			} else {
				_, err := io.Copy(file, resp.Body)
				if err != nil {
					util.Stderr(err)
					util.SubtleHighlight(msg)
				}
			}
		}
	} else {
		// generate main file
		source, _ = assetsFS.ReadFile("assets/main.go.tpl")
		content = util.ApplyTemplate(source, map[string]string{})
		os.MkdirAll(filepath.Join(abspath, "cmd", appname), os.ModePerm)
		util.WriteTextFile(filepath.Join(abspath, "cmd", appname, "main.go"), content, true)

		// // generate main test file
		// source, _ = assetsFS.ReadFile("assets/main_test.go.tpl")
		// content = util.ApplyTemplate(source, map[string]string{
		// 	"Name":        appname,
		// 	"Version":     "0.0.1",
		// 	"Description": description,
		// })
		// util.WriteTextFile(filepath.Join(abspath, "main_test.go"), content, true)
	}

	// if strings.Contains(createtype, "command") {
	// 	if strings.Contains(createtype, "module") {
	// 		source, _ = assetsFS.ReadFile("assets/cmd.module.go.tpl")
	// 	} else {
	// 		source, _ = assetsFS.ReadFile("assets/cmd.go.tpl")
	// 	}
	// 	content = util.ApplyTemplate(source, map[string]string{
	// 		"Module":      modulename,
	// 		"ModuleAlias": appname,
	// 	})

	// 	if strings.Contains(createtype, "module") {
	// 		buildcmd = true
	// 		util.BailOnError(os.MkdirAll(filepath.Join(abspath, "cmd"), os.ModePerm))
	// 		util.WriteTextFile(filepath.Join(abspath, "cmd", "main.go"), content, true)
	// 	} else {
	// 		util.WriteTextFile(filepath.Join(abspath, appname+".go"), content, true)
	// 	}

	// 	if strings.Contains(createtype, "multi") {
	// 		content = strings.ReplaceAll(content, "\"You are %d years old.", "person +\" is %d years old.")
	// 		util.BailOnError(os.MkdirAll(filepath.Join(abspath, "alt_cmd"), os.ModePerm))
	// 		util.WriteTextFile(filepath.Join(abspath, "alt_cmd", "main.go"), content, true)
	// 	}
	// }

	// if strings.Contains(createtype, "supported") || strings.Contains(createtype, "multipackage") {
	// 	source, _ = assetsFS.ReadFile("assets/support.go.tpl")
	// 	content = util.ApplyTemplate(source, map[string]string{
	// 		"Module":     modulename,
	// 		"ModuleName": "submodule",
	// 		"FuncName":   "TranslatedGreeting",
	// 		"Greeting":   "Guten Tag",
	// 	})
	// 	util.Command("mkdir -p", filepath.Join(abspath, "internal", "submodule"))
	// 	util.BailOnError(os.MkdirAll(filepath.Join(abspath, "internal", "submodule"), os.ModePerm))
	// 	util.WriteTextFile(filepath.Join(abspath, "internal", "submodule", "submodule.go"), content, true)

	// 	// Add internal module example
	// 	if strings.Contains(createtype, "multipackage") {
	// 		mods := []string{"module_a", "submodule_b"}
	// 		for _, mod := range mods {
	// 			content = util.ApplyTemplate(source, map[string]string{
	// 				"Module":     modulename,
	// 				"ModuleName": mod,
	// 				"FuncName":   "GreetingFrom_" + mod,
	// 				"Greeting":   mod + " says hello",
	// 			})
	// 			util.Command("mkdir -p", filepath.Join(abspath, mod))
	// 			util.BailOnError(os.MkdirAll(filepath.Join(abspath, mod), os.ModePerm))
	// 			util.WriteTextFile(filepath.Join(abspath, mod, mod+".go"), strings.Replace(content, "/internal/", "/", 1), true)
	// 		}
	// 	}
	// }

	// if strings.Contains(createtype, "server") {
	// 	source, _ = assetsFS.ReadFile("assets/server.go.tpl")
	// 	content = util.ApplyTemplate(source, map[string]string{
	// 		"Module": modulename,
	// 	})
	// 	util.WriteTextFile(filepath.Join(abspath, "cmd", appname+".go"), content, true)

	// 	source, _ = assetsFS.ReadFile("assets/server_module.go.tpl")
	// 	content = util.ApplyTemplate(source, map[string]string{})
	// 	util.WriteTextFile(filepath.Join(abspath, "greeting", "greeting.go"), content, true)
	// }

	// generate manifest.json
	manifesttpl := "assets/manifest.tpl"
	if createtype == "module" {
		manifesttpl = "assets/manifest.module.tpl"
	}
	source, _ = assetsFS.ReadFile(manifesttpl)
	if license == "custom" {
		license = "SEE LICENSE IN LICENSE"
	}
	mainfile := strings.ToLower(appname) + ".go"
	if createtype == "command" {
		mainfile = "main.go"
	}
	wasm := "false"
	if createtype == "wasm" {
		wasm = "true"
	}
	content = util.ApplyTemplate(source, map[string]string{
		"Author":      author,
		"Name":        appname,
		"Description": strings.ReplaceAll(description, "\"", "\\\""),
		"License":     license,
		"Main":        mainfile,
		"Version":     "0.0.1",
		"Module":      mod,
		"WASM":        wasm,
	})
	util.WriteTextFile(filepath.Join(abspath, "manifest.json"), content, true)

	// Initialize go work
	err = util.Run("go work init", abspath)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			util.Stderr(err)
		}
	}

	err = util.Run("go work use .", abspath)
	if err != nil {
		util.Stderr(err)
	}

	// generate git repo
	if usegit {
		// generate .gitignore
		source, _ = assetsFS.ReadFile("assets/gitignore.tpl")
		content = util.ApplyTemplate(source, map[string]string{
			"Name": appname,
		})
		util.WriteTextFile(filepath.Join(abspath, ".gitignore"), content, true)

		// Initialize local git
		util.Run("git init", abspath)

		// Add remote repo if specified
		if remoteorigin {
			err := util.Run("git add remote origin "+remoterepo, abspath)
			if err != nil {
				util.Stderr(err)
			}
		}
	}

	return nil
}

func addFileToZip(zipWriter *zip.Writer, abspath string, file string) error {
	// Open the file or directory to be added to the zip archive
	fileToZip, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	if info.IsDir() {
		// If it's a directory, add its contents recursively
		files, err := ioutil.ReadDir(file)
		if err != nil {
			return err
		}

		for _, f := range files {
			path := filepath.Join(file, f.Name())
			if err := addFileToZip(zipWriter, abspath, path); err != nil {
				return err
			}
		}
	} else {
		// If it's a file, create a header and copy it to the zip archive
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Specify the name of the file within the zip archive
		header.Name = strings.TrimPrefix(filepath.ToSlash(file), filepath.ToSlash(abspath)+"/")

		// Create a writer for the file within the zip archive
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// Copy the file data to the zip archive
		_, err = io.Copy(writer, fileToZip)
		if err != nil {
			return err
		}
	}

	return nil
}
