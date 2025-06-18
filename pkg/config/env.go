// Package config provides environment variable loading utilities with type conversion and defaults.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// EnvLoader provides methods for loading and parsing environment variables with default values.
// It supports type conversion for common data types used in application configuration.
type EnvLoader struct{}

// Env is a global instance of EnvLoader for convenient access to environment variable loading.
var Env = &EnvLoader{}

// String loads a string environment variable with a default value fallback.
// Returns the environment variable value if set, otherwise returns the default value.
func (e *EnvLoader) String(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Int loads an integer environment variable with a default value fallback.
// Attempts to parse the environment variable as an integer, returns default on failure.
func (e *EnvLoader) Int(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// Int64 loads a 64-bit integer environment variable with a default value fallback.
// Attempts to parse the environment variable as an int64, returns default on failure.
func (e *EnvLoader) Int64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if int64Value, err := strconv.ParseInt(value, 10, 64); err == nil {
			return int64Value
		}
	}
	return defaultValue
}

// Float64 loads a float64 environment variable with a default value fallback.
// Attempts to parse the environment variable as a float64, returns default on failure.
func (e *EnvLoader) Float64(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// Bool loads a boolean environment variable with a default value fallback.
// Accepts common boolean representations (true/false, 1/0, etc.), returns default on failure.
func (e *EnvLoader) Bool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// Duration loads a time.Duration environment variable with a default value fallback.
// Accepts duration strings like "30s", "5m", "2h", returns default on parse failure.
func (e *EnvLoader) Duration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// StringRequired loads a required string environment variable, returning an error if not set.
// Use this for critical configuration values that must be provided.
func (e *EnvLoader) StringRequired(key string) (string, error) {
	if value := os.Getenv(key); value != "" {
		return value, nil
	}
	return "", fmt.Errorf("required environment variable %s is not set", key)
}

// IntRequired loads a required integer environment variable, returning an error if not set or invalid.
// Use this for critical numeric configuration values that must be provided.
func (e *EnvLoader) IntRequired(key string) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return 0, fmt.Errorf("required environment variable %s is not set", key)
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid integer value for %s: %s", key, value)
	}
	return intValue, nil
}
