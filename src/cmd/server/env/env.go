package env

import (
	"fmt"
	"os"
)

// Returns environment variable value, or a default value specified in the parameters
func Get(name string, defaultValue string) string {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue
	}

	return value
}

// Returns environment variable value or error if the variable is empty
func GetNeeded(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("Environment variable with the name %s doen't exist, or is empty.", name)
	}

	return value, nil
}
