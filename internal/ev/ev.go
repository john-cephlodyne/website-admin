package ev

import (
	"fmt"
	"os"
	"sync"
)

var runningInCloud = true

// Use a struct to hold the cache and its mutex together.
var envCache = struct {
	mu    sync.Mutex
	items map[string]string
}{
	items: make(map[string]string),
}

// Get retrieves an environment variable. It returns an error if the variable is not set.
func Get(varName string) (string, error) {
	envCache.mu.Lock()
	defer envCache.mu.Unlock()

	value, ok := envCache.items[varName]
	if ok {
		return value, nil // Found in cache
	}

	value = os.Getenv(varName)
	if value == "" {
		return "", fmt.Errorf("environment variable %q is not set", varName)
	}

	envCache.items[varName] = value
	return value, nil
}

// IsRunningInCloud returns the value of the compile-time flag.
func IsRunningInCloud() bool {
	return runningInCloud
}
