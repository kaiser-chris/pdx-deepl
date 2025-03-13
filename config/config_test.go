package config

import (
	"fmt"
	"net/url"
	"testing"
)

func TestExistingEnvString(t *testing.T) {
	expected := "value"
	t.Setenv("TEST", expected)
	result := RequireStringEnv("TEST")
	if result != expected {
		t.Logf("Environment variable is not the expected value: %s != %s", expected, result)
		t.Fail()
	}
}

func TestExistingEnvBool(t *testing.T) {
	expected := true
	t.Setenv("TEST_BOOL", fmt.Sprint(expected))
	result := RequireBoolEnv("TEST_BOOL")
	if result != expected {
		t.Logf("Environment variable is not the expected value: %s != %s", fmt.Sprint(expected), fmt.Sprint(result))
		t.Fail()
	}
}

func TestExistingEnvInt(t *testing.T) {
	expected := 42
	t.Setenv("TEST_INT", fmt.Sprint(expected))
	result := RequireIntEnv("TEST_INT")
	if result != expected {
		t.Logf("Environment variable is not the expected value: %s != %s", fmt.Sprint(expected), fmt.Sprint(result))
		t.Fail()
	}
}

func TestExistingEnvUrl(t *testing.T) {
	expected, _ := url.Parse("localhost:8000/test")
	t.Setenv("TEST_URL", fmt.Sprint(expected))
	result := RequireUrlEnv("TEST_URL")
	if fmt.Sprint(result) != fmt.Sprint(expected) {
		t.Logf("Environment variable is not the expected value: %s != %s", fmt.Sprint(expected), fmt.Sprint(result))
		t.Fail()
	}
}
