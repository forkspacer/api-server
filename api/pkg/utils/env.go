package utils

import (
	"errors"
	"fmt"
	"os"
)

var (
	ErrEnvNotFound = errors.New("environment variable not found")
)

// GetEnv retrieves the environment variable named by the key envName,
// and attempts to parse its value into the specified type T.
//
// It returns the parsed value of type T and a nil error on success.
// If the environment variable is not set, it returns a zero value of type T and ErrEnvNotFound.
// If parsing fails, it returns a zero value of type T and a formatted error describing the failure.
func GetEnv[T ParseStringSupportTypes](envName string) (T, error) {
	var zero T

	envStr := os.Getenv(envName)
	if envStr == "" {
		return zero, ErrEnvNotFound
	}

	parsedEnv, err := ParseString[T](envStr)
	if err != nil {
		return zero, fmt.Errorf("failed to parse environment variable %s: %s", envName, err)
	}

	return parsedEnv, nil
}

// GetEnvOr retrieves the environment variable named by the key envName,
// and attempts to parse its value into the specified type T.
//
// It returns the parsed value of type T and a nil error on success.
// If the environment variable is not set, it returns the provided default value
// and ErrEnvNotFound. If parsing fails, it returns a zero value of type T and a formatted error describing the failure.
func GetEnvOr[T ParseStringSupportTypes](envName string, defaultValue T) (T, error) {
	var zero T

	envStr := os.Getenv(envName)
	if envStr == "" {
		return defaultValue, ErrEnvNotFound
	}

	parsedEnv, err := ParseString[T](envStr)
	if err != nil {
		return zero, fmt.Errorf("failed to parse environment variable %s: %s", envName, err)
	}

	return parsedEnv, nil
}
