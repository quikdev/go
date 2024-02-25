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

func New() *Config {
	cfgfile := "manifest.json"
	data, err := readJSON(cfgfile)
	if err != nil {
		cfgfile = "package.json"
		data, err = readJSON(cfgfile)
		if err != nil {
			var empty string
			cfgfile = empty
			// fmt.Println(err)
		}
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

		magenta := color.New(color.FgMagenta, color.Bold).SprintFunc()
		fmt.Printf(magenta("\nUsing %s configuration\n\n"), file)
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
