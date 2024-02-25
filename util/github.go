package util

import (
	"errors"
	"os/exec"
	"strings"
)

// Function to get GitHub handle and email from global Git config
func GetGitHubInfo() (handle, email string, err error) {
	// Get email from global Git config
	email, err = getGitConfigValue("user.email")
	if err != nil {
		return "", "", errors.New("failed to get GitHub email: " + err.Error())
	}

	// Without a repository, we can't reliably determine the GitHub handle
	return "", email, nil
}

// Function to get a value from global Git config
func getGitConfigValue(key string) (string, error) {
	cmd := exec.Command("git", "config", "--global", "--get", key)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func ExtractGitHubHandleFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) >= 2 && strings.HasSuffix(parts[1], "users.noreply.github.com") {
		usernameParts := strings.Split(parts[0], "+")
		if len(usernameParts) > 1 {
			return usernameParts[len(usernameParts)-1]
		}
		return parts[0]
	}
	return ""
}
