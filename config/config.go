package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

type Config struct {
	data    map[string]interface{}
	cfgfile string
}

var jsonfiles = []string{"manifest"}
var warned = false

func New() *Config {
	cfgfile := "manifest.json"
	data, err := readJSON(cfgfile)
	if err != nil {
		var empty string
		cfgfile = empty
	}

	return &Config{data: data, cfgfile: cfgfile}
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
