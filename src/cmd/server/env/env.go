package env

import (
	"fmt"
	"os"
)

// Get vrátí hodnotu proměnné prostředí, nebo výchozí hodnotu pokud není nastavena.
func Get(name string, defaultValue string) string {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue
	}

	return value
}

// GetNeeded vrátí hodnotu proměnné prostředí nebo chybu pokud je prázdná.
func GetNeeded(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("proměnná prostředí %s neexistuje nebo je prázdná", name)
	}

	return value, nil
}
