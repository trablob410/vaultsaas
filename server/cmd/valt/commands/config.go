package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds CLI configuration.
type Config struct {
	Token  string `json:"token"`
	APIURL string `json:"api_url"`
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".valt", "config.json")
}

// LoadConfig reads config from ~/.valt/config.json or env vars.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		Token:  os.Getenv("VALT_TOKEN"),
		APIURL: os.Getenv("VALT_API_URL"),
	}
	if cfg.APIURL == "" {
		cfg.APIURL = "http://localhost:8080/api/v1"
	}

	path := configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	var fileCfg Config
	if err := json.Unmarshal(data, &fileCfg); err != nil {
		return nil, err
	}

	// Env vars take precedence over file
	if cfg.Token == "" {
		cfg.Token = fileCfg.Token
	}
	if cfg.APIURL == "http://localhost:8080/api/v1" && fileCfg.APIURL != "" {
		cfg.APIURL = fileCfg.APIURL
	}

	return cfg, nil
}

// SaveConfig writes config to ~/.valt/config.json.
func SaveConfig(cfg *Config) error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
