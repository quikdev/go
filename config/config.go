package config

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/peterbourgon/mergemap"
	"github.com/quikdev/go/util"
)

type Config struct {
	data    map[string]interface{}
	cfgfile string
	exists  bool
}

var jsonfiles = []string{"manifest"}
var warned = false
var warnedprofiles = false

func New(profiles ...string) *Config {
	cfgfile := "manifest.json"
	exists := false
	data, err := readJSON(cfgfile)
	if err != nil {
		var emptystr string
		cfgfile = emptystr
	} else {
		exists = true

		// If no profiles are specified and the default_profile exists in the
		// manifest, apply it.
		if len(profiles) == 0 {
			if p, exists := data["default_profiles"]; exists {
				defaults := p.([]interface{})
				for _, value := range defaults {
					profiles = append(profiles, value.(string))
				}
			}
		}

		// Apply OS-level profiles
		if prof, exists := data["profile"]; exists {
			os := strings.ToLower(runtime.GOOS)
			if _, exists := prof.(map[string]interface{})[os]; exists {
				profiles = append(profiles, os)
			}
		}

		// Check profiles, then apply the specified profile(s)
		// to the data by merging the JSON attributes. This should
		// provide the overrides needed to support this feature.
		if len(profiles) > 0 {
			used := []string{}
			availableprofiles := []string{}
			if profileData, exists := data["profile"]; exists {
				for name, profile := range profileData.(map[string]interface{}) {
					availableprofiles = append(availableprofiles, name)
					if util.IndexOf[string](profiles, name) >= 0 {
						_, err := json.MarshalIndent(profile.(map[string]interface{}), "", "  ")
						if err != nil {
							util.Stderr(err, true)
						}
						data = mergemap.Merge(data, profile.(map[string]interface{}))
						used = append(used, name)
					}
				}
			}

			// Notify user if a missing profile is specified
			if len(used) == 0 {
				if len(availableprofiles) == 0 {
					plural := ""
					if len(profiles) != 1 {
						plural = "s"
					}
					util.Stderr(fmt.Sprintf(`%s profile%s not found (no profiles available in manifest.json)`, strings.Join(profiles, "/"), plural), true)
				}

				missing := []string{}
				for _, name := range profiles {
					if util.IndexOf[string](availableprofiles, name) < 0 {
						missing = append(missing, name)
					}
				}

				plural := ""
				if len(missing) != 1 {
					plural = "s"
				}
				util.Stderr(fmt.Sprintf(`%s profile%s not found in manifest.json - please use one/more of the following: %s (or create the missing profile%s)`, strings.Join(profiles, "/"), plural, strings.Join(availableprofiles, ", "), plural), true)
			}

			if !warnedprofiles && len(profiles) > 0 && os.Args[1] != "exec" {
				magenta := color.New(color.FgMagenta, color.Faint, color.Italic).SprintFunc()
				dim := color.New(color.Faint).SprintFunc()
				plural := ""
				if len(profiles) != 1 {
					plural = "s"
				}
				util.Stdout(fmt.Sprintf(" with %s"+dim(" profile%s applied\n"), magenta(strings.Join(used, "+")), plural))
				warnedprofiles = true
			}
		}

		delete(data, "profile")
		delete(data, "default_profiles")
	}

	// Adds an extra break after manifest notification
	fmt.Println("")

	return &Config{data: data, cfgfile: cfgfile, exists: exists}
}

func readJSON(file string) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	fileContents, err := os.ReadFile(file)
	if err != nil {
		if strings.Contains(err.Error(), "cannot find the file") {
			if !warned {
				util.Stdout("\n# no manifest available\n\n")
				warned = true
			}
		}

		return data, err
	} else {
		err = json.Unmarshal(fileContents, &data)
		if err != nil {
			if syntaxErr, ok := err.(*json.SyntaxError); ok {
				// Calculate the line number based on the byte offset
				lineNumber := strings.Count(string(fileContents[:syntaxErr.Offset]), "\n") + 1
				util.Stderr(fmt.Errorf("error reading manifest: %v\n  at %s:%d", err, file, lineNumber), true)
			}

			util.Stderr(fmt.Errorf("error reading manifest: %s", err.Error()), true)
			// return data, err
		}

		if !warned {
			magenta := color.New(color.FgMagenta, color.Faint).SprintFunc()
			dim := color.New(color.Faint).SprintFunc()
			fmt.Printf(dim("\n# using "+magenta("%s")+dim(" configuration")), file)
			warned = true
		}
	}

	return data, nil
}

var empty interface{}

func (cfg *Config) ManifestExists() bool {
	return cfg.exists
}

func (cfg *Config) Get(name string) (interface{}, bool) {
	if strings.Contains(name, ".") {
		var data interface{}
		data = cfg.data
		parts := strings.Split(name, ".")
		for i, part := range parts {
			if value, exists := data.(map[string]interface{})[part]; exists {
				if len(parts) == (i + 1) {
					return value, true
				}

				data = value
			} else {
				return empty, false
			}
		}
	} else {
		if value, exists := cfg.data[name]; exists {
			return value, true
		}
	}

	return empty, false
}

func (cfg *Config) Data() map[string]interface{} {
	return cfg.data
}

func (cfg *Config) File() string {
	return cfg.cfgfile
}

func (cfg *Config) Raw() ([]byte, error) {
	return os.ReadFile(cfg.cfgfile)
}

func (cfg *Config) GetEnvVars() map[string]string {
	if cfg.data == nil {
		data, _ := readJSON(cfg.cfgfile)
		cfg.data = data
	}

	result := make(map[string]string)

	if env, exists := cfg.Get("env"); exists {
		for key, value := range env.(map[string]interface{}) {
			for _, prefix := range jsonfiles {
				if strings.HasPrefix(key, prefix+".") {
					if subkey, keyexists := cfg.Get(strings.Replace(key, prefix+".", "", 1)); keyexists {
						value = subkey
						break
					}
				}
			}

			result[key] = value.(string)
		}
	}

	return result
}

func (cfg *Config) GetEnvVarList() []string {
	result := []string{}
	for key, value := range cfg.GetEnvVars() {
		result = append(result, fmt.Sprintf("%s=%s", key, value))
	}

	return result
}
