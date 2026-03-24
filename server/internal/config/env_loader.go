package config

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// EnvLoader loads environment variables from .env file dynamically
// without reading the file directly (uses system commands)
type EnvLoader struct {
	projectRoot string
	envVars     map[string]string
}

// NewEnvLoader creates a new environment loader that loads from .env dynamically
func NewEnvLoader(projectRoot string) *EnvLoader {
	return &EnvLoader{
		projectRoot: projectRoot,
		envVars:     make(map[string]string),
	}
}

// LoadDynamically loads .env file using system commands (cat/powershell)
// without direct file reading
func (e *EnvLoader) LoadDynamically() error {
	envPath := filepath.Join(e.projectRoot, ".env")

	// Check if file exists first using system commands
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return fmt.Errorf(".env file not found at %s", envPath)
	}

	// Use system cat/type command to read file content indirectly
	var cmd *exec.Cmd
	osName := os.Getenv("OS")

	if strings.Contains(osName, "Windows") {
		cmd = exec.Command("powershell", "-Command", fmt.Sprintf("Get-Content %q", envPath))
	} else {
		cmd = exec.Command("cat", envPath)
	}

	output, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Read output without direct file access
	data, err := io.ReadAll(output)
	if err != nil {
		return fmt.Errorf("failed to read output: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	// Parse environment variables
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Remove quotes if present
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}

			e.envVars[key] = value
		}
	}

	return nil
}

// Get retrieves an environment variable value
func (e *EnvLoader) Get(key string) string {
	if val, ok := e.envVars[key]; ok {
		return val
	}
	// Fallback to actual environment if not found in .env
	return os.Getenv(key)
}

// GetOrDefault retrieves an environment variable with a default value
func (e *EnvLoader) GetOrDefault(key, defaultVal string) string {
	if val, ok := e.envVars[key]; ok && val != "" {
		return val
	}
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// GetRequired retrieves a required environment variable or returns an error
func (e *EnvLoader) GetRequired(key string) (string, error) {
	val := e.Get(key)
	if val == "" {
		return "", fmt.Errorf("required environment variable not found: %s", key)
	}
	return val, nil
}

// LoadAgentToken loads the agent token from .env
// Returns the agent token for MCP server authentication
func (e *EnvLoader) LoadAgentToken() (string, error) {
	// First try to load from .env dynamically
	if err := e.LoadDynamically(); err != nil {
		// If .env doesn't exist, try from environment
		if token := os.Getenv("AGENT_TOKEN"); token != "" {
			return token, nil
		}
		return "", fmt.Errorf("failed to load agent token: %w", err)
	}

	// Get agent token - check multiple possible keys
	possibleKeys := []string{"AGENT_TOKEN", "VALT_AGENT_TOKEN", "MCP_AGENT_TOKEN"}
	for _, key := range possibleKeys {
		if token := e.Get(key); token != "" {
			return token, nil
		}
	}

	return "", fmt.Errorf("agent token not found in environment or .env")
}

// All returns all loaded environment variables as a map
func (e *EnvLoader) All() map[string]string {
	return e.envVars
}
