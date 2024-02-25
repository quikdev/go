package util

import (
	"log"
	"os/exec"
	"regexp"
)

func GetCurrentGoVersion() string {
	cmd := exec.Command("go", "version")

	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("Error identifying go version: %v", err)
	}

	versionRegex := regexp.MustCompile(`go version go(\d+\.\d+)`)
	match := versionRegex.FindStringSubmatch(string(output))
	if len(match) < 2 {
		log.Fatalf("Error parsing Go version number")
	}

	version := match[1]
	return version
}
