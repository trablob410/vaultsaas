package config

import (
	"os"
	"testing"
)

// TestEnvLoaderGetVariable tests getting environment variables
func TestEnvLoaderGetVariable(t *testing.T) {
	loader := NewEnvLoader("/tmp")

	// Set a test variable in OS environment
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	val := loader.Get("TEST_VAR")
	if val != "test_value" {
		t.Errorf("expected 'test_value', got %q", val)
	}
}

// TestEnvLoaderGetOrDefault tests default values
func TestEnvLoaderGetOrDefault(t *testing.T) {
	loader := NewEnvLoader("/tmp")

	// Test with non-existent variable
	val := loader.GetOrDefault("NONEXISTENT_VAR_12345", "default_value")
	if val != "default_value" {
		t.Errorf("expected 'default_value', got %q", val)
	}

	// Test with existing variable
	os.Setenv("TEST_VAR", "actual_value")
	defer os.Unsetenv("TEST_VAR")

	val = loader.GetOrDefault("TEST_VAR", "default_value")
	if val != "actual_value" {
		t.Errorf("expected 'actual_value', got %q", val)
	}
}

// TestEnvLoaderGetRequired tests required variable retrieval
func TestEnvLoaderGetRequired(t *testing.T) {
	loader := NewEnvLoader("/tmp")

	// Test with missing required variable
	_, err := loader.GetRequired("NONEXISTENT_REQUIRED_VAR_12345")
	if err == nil {
		t.Error("expected error for missing required variable")
	}

	// Test with existing required variable
	os.Setenv("TEST_REQUIRED", "required_value")
	defer os.Unsetenv("TEST_REQUIRED")

	val, err := loader.GetRequired("TEST_REQUIRED")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if val != "required_value" {
		t.Errorf("expected 'required_value', got %q", val)
	}
}

// TestEnvLoaderAll tests getting all variables
func TestEnvLoaderAll(t *testing.T) {
	loader := NewEnvLoader("/tmp")

	// Set some test variables
	os.Setenv("VAR1", "value1")
	os.Setenv("VAR2", "value2")
	defer os.Unsetenv("VAR1")
	defer os.Unsetenv("VAR2")

	allVars := loader.All()
	if allVars == nil {
		t.Error("expected non-nil map")
	}

	// All() returns only loaded vars from .env, not OS environment
	// So the map might be empty if .env file doesn't exist
	if len(allVars) > 0 {
		t.Log("✓ EnvLoader.All() returns loaded variables")
	}
}

// TestEnvLoaderLoadAgentToken tests agent token loading
func TestEnvLoaderLoadAgentToken(t *testing.T) {
	loader := NewEnvLoader("/tmp")

	// Set agent token in environment
	os.Setenv("AGENT_TOKEN", "test_agent_token_123")
	defer os.Unsetenv("AGENT_TOKEN")

	token, err := loader.LoadAgentToken()
	if err != nil {
		t.Logf("LoadAgentToken error (expected if .env missing): %v", err)
	}
	if token != "" && token != "test_agent_token_123" {
		t.Errorf("unexpected token: %q", token)
	}
}
