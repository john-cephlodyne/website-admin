package ev

import (
	"fmt"
	"os"
	"strings"
)

var cloudFlag = "true"

// Get retrieves a standard environment variable.
func Get(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("environment variable %q is not set", name)
	}
	return value, nil
}

// GetSecret retrieves a secret strictly from the provided absolute file path.
func GetSecret(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("secret file not found or unreadable at %q: %w", filePath, err)
	}

	// Still keep this! Trailing newlines will ruin your day.
	return strings.TrimSpace(string(data)), nil
}

// IsRunningInCloud returns true if the compile-time flag is set to "true".
func IsRunningInCloud() bool {
	return cloudFlag == "true"
}
