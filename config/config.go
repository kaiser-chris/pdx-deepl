package config

import (
	"net/url"
	"os"
	"strconv"

	"bahmut.de/pdx-deepl/util/logging"
)

var Directory = "."

// Defaults to an empty string
func OptionalStringEnv(key string) string {
	return os.Getenv(key)
}

// Defaults to false
func OptionalBoolEnv(key string) bool {
	value := OptionalStringEnv(key)
	if value == "" {
		return false
	}
	return OptionalBoolEnv(key)
}

// Defaults to 0
func OptionalIntEnv(key string) int {
	value := OptionalStringEnv(key)
	if value == "" {
		return 0
	}
	return RequireIntEnv(key)
}

// Defaults to nil
func OptionalUrlEnv(key string) *url.URL {
	value := OptionalStringEnv(key)
	if value == "" {
		return nil
	}
	return RequireUrlEnv(key)
}

func RequireStringEnv(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		logging.Fatalf("Could not find required environment variable: %s", key)
	}
	return value
}

func RequireUrlEnv(key string) *url.URL {
	value := RequireStringEnv(key)
	urlValue, err := url.Parse(value)
	if err != nil {
		logging.Fatalf("Could not parse value (%s) from environment variable %s: Expected URL value", value, key)
	}
	return urlValue
}

func RequireBoolEnv(key string) bool {
	value := RequireStringEnv(key)
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		logging.Fatalf("Could not parse value (%s) from environment variable %s: Expected boolean value", value, key)
	}
	return boolValue
}

func RequireIntEnv(key string) int {
	value := RequireStringEnv(key)
	intValue, err := strconv.Atoi(value)
	if err != nil {
		logging.Fatalf("Could not parse value (%s) from environment variable %s: Expected int value", value, key)
	}
	return intValue
}
