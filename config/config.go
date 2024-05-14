package config

import (
	"encoding/json"
	"fmt"
	"os"
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

func New(profiles ...string) *Config {
	cfgfile := "manifest.json"
	exists := false
	data, err := readJSON(cfgfile)
	if err != nil {
		var emptystr string
		cfgfile = emptystr
	} else {
		exists = true
	}

	// If no profiles are specified and the default_profile exists in the
	// manifest, apply it.
	if len(profiles) == 0 {
		if p, exists := data["default_profile"]; exists {
			profiles = []string{p.(string)}
		}
	}

	// Check profiles, then apply the specified profile(s)
	// to the data by merging the JSON attributes. This should
	// provide the overrides needed to support this feature.
	if len(profiles) > 0 {
		if profileData, exists := data["profile"]; exists {
			for name, profile := range profileData.(map[string]interface{}) {
				if util.IndexOf[string](profiles, name) >= 0 {
					data = mergemap.Merge(data, profile.(map[string]interface{}))
				}
			}
		}
	}

	delete(data, "profile")

	return &Config{data: data, cfgfile: cfgfile, exists: exists}
}

func readJSON(file string) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	fileContents, err := os.ReadFile(file)
	if err != nil {
		return data, err
	} else {
		err = json.Unmarshal(fileContents, &data)
		if err != nil {
			return data, err
		}

		if !warned {
			magenta := color.New(color.FgMagenta, color.Faint).SprintFunc()
			dim := color.New(color.Faint).SprintFunc()
			fmt.Printf(dim("\n# using "+magenta("%s")+dim(" configuration\n")), file)
			fmt.Println("")
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
