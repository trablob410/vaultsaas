package config

import (
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Point config path to temp dir with no file → should return defaults
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIURL == "" {
		t.Error("APIURL should have default")
	}
}
