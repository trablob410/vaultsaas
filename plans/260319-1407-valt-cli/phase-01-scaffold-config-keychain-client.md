# Phase 01: Scaffold + Config + Keychain + API Client

**Priority:** P0 — foundation for all CLI commands
**Status:** COMPLETED — 2026-03-19

## Files

| File | Change |
|------|--------|
| `valt-cli/go.mod` | Create |
| `valt-cli/cmd/root.go` | Create |
| `valt-cli/internal/config/config.go` | Create |
| `valt-cli/internal/keychain/keychain.go` | Create |
| `valt-cli/internal/api/client.go` | Create |

---

## Task 1: Init Go module

- [ ] Create module:

```bash
cd D:/vaultsaas/valt-cli
go mod init github.com/valt-dev/valt/valt-cli
go get github.com/spf13/cobra@v1.8.0
go get github.com/zalando/go-keyring@v0.2.4
go get github.com/BurntSushi/toml@v1.3.2
```

- [ ] Commit:
```bash
git add valt-cli/go.mod valt-cli/go.sum
git commit -m "chore(valt-cli): init Go module with cobra, go-keyring, toml deps"
```

---

## Task 2: internal/config/config.go

**Files:**
- Create: `valt-cli/internal/config/config.go`

- [ ] Write:

```go
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds CLI configuration persisted to ~/.valt/config.toml.
type Config struct {
	APIURL    string `toml:"api_url"`
	ProjectID string `toml:"project_id,omitempty"`
}

func defaultAPIURL() string { return "https://api.valt.dev" }

// Load reads config from ~/.valt/config.toml. Returns defaults if file doesn't exist.
func Load() (*Config, error) {
	cfg := &Config{APIURL: defaultAPIURL()}
	path, err := configPath()
	if err != nil {
		return cfg, nil // best-effort: return defaults
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	// Env overrides
	if v := os.Getenv("VALT_API_URL"); v != "" { cfg.APIURL = v }
	if v := os.Getenv("VALT_PROJECT_ID"); v != "" { cfg.ProjectID = v }
	return cfg, nil
}

// Save writes config to ~/.valt/config.toml, creating the directory if needed.
func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("opening config file: %w", err)
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("finding home dir: %w", err)
	}
	return filepath.Join(home, ".valt", "config.toml"), nil
}
```

- [ ] Write tests:

```go
// valt-cli/internal/config/config_test.go
func TestLoadDefaults(t *testing.T) {
    // Point config path to temp dir with no file → should return defaults
    cfg, err := Load()
    if err != nil { t.Fatal(err) }
    if cfg.APIURL == "" { t.Error("APIURL should have default") }
}
```

- [ ] Run: `cd valt-cli && go test ./internal/config/...`

- [ ] Commit:
```bash
git add valt-cli/internal/config/
git commit -m "feat(valt-cli): add config load/save for ~/.valt/config.toml"
```

---

## Task 3: internal/keychain/keychain.go

**Files:**
- Create: `valt-cli/internal/keychain/keychain.go`

- [ ] Write (mirrors `mcp-server/src/keychain.rs` pattern in Go):

```go
package keychain

import (
	"fmt"
	"os"

	"github.com/zalando/go-keyring"
)

const service = "valt-cli"
const userKey = "auth-token"

// GetToken returns the stored auth token, checking env var first.
func GetToken() (string, error) {
	if v := os.Getenv("VALT_AUTH_TOKEN"); v != "" {
		return v, nil
	}
	token, err := keyring.Get(service, userKey)
	if err != nil {
		return "", nil // not found = not logged in
	}
	return token, nil
}

// SetToken stores the auth token in the OS keychain.
func SetToken(token string) error {
	if err := keyring.Set(service, userKey, token); err != nil {
		return fmt.Errorf("storing token in keychain: %w", err)
	}
	return nil
}

// DeleteToken removes the stored token.
func DeleteToken() error {
	return keyring.Delete(service, userKey)
}
```

- [ ] Run: `cd valt-cli && go build ./internal/keychain/...`

- [ ] Commit:
```bash
git add valt-cli/internal/keychain/
git commit -m "feat(valt-cli): add OS keychain token storage"
```

---

## Task 4: internal/api/client.go

**Files:**
- Create: `valt-cli/internal/api/client.go`

- [ ] Write HTTP client:

```go
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Client talks to the Valt REST API.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func New(baseURL, token string) *Client {
	return &Client{baseURL: baseURL, token: token, http: &http.Client{}}
}

func (c *Client) do(ctx context.Context, method, path string, body interface{}, out interface{}) error {
	var bodyReader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil { return err }
		bodyReader = bytes.NewReader(b)
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+"/api/v1"+path, bodyReader)
	if err != nil { return err }
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil { return fmt.Errorf("request failed: %w", err) }
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errBody struct{ Error string `json:"error"` }
		json.NewDecoder(resp.Body).Decode(&errBody)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, errBody.Error)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

// Get performs GET /api/v1{path} and decodes JSON response into out.
func (c *Client) Get(ctx context.Context, path string, out interface{}) error {
	return c.do(ctx, "GET", path, nil, out)
}

// Post performs POST /api/v1{path} with body, decodes into out.
func (c *Client) Post(ctx context.Context, path string, body, out interface{}) error {
	return c.do(ctx, "POST", path, body, out)
}
```

- [ ] Write test (mock HTTP server):

```go
func TestClientGet_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer srv.Close()
	c := New(srv.URL, "test-token")
	var out map[string]string
	if err := c.Get(context.Background(), "/health", &out); err != nil {
		t.Fatal(err)
	}
	if out["status"] != "ok" { t.Error("unexpected response") }
}
```

- [ ] Run: `cd valt-cli && go test ./internal/api/...`

- [ ] Commit:
```bash
git add valt-cli/internal/api/
git commit -m "feat(valt-cli): add REST API client with Bearer token auth"
```

---

## Task 5: cmd/root.go — Cobra root

**Files:**
- Create: `valt-cli/cmd/root.go`
- Create: `valt-cli/main.go`

- [ ] Write root command:

```go
// valt-cli/cmd/root.go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/valt-dev/valt/valt-cli/internal/api"
	"github.com/valt-dev/valt/valt-cli/internal/config"
	"github.com/valt-dev/valt/valt-cli/internal/keychain"
)

var rootCmd = &cobra.Command{
	Use:     "valt",
	Short:   "Valt — secret vault CLI",
	Version: "0.1.0",
}

// MustClient returns a configured API client, exits if not authenticated.
func MustClient() *api.Client {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading config:", err)
		os.Exit(1)
	}
	token, err := keychain.GetToken()
	if err != nil || token == "" {
		fmt.Fprintln(os.Stderr, "Not logged in. Run: valt setup")
		os.Exit(1)
	}
	return api.New(cfg.APIURL, token)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

```go
// valt-cli/main.go
package main

import "github.com/valt-dev/valt/valt-cli/cmd"

func main() { cmd.Execute() }
```

- [ ] Run: `cd valt-cli && go build ./... && ./valt-cli --version`
  Expected: `valt version 0.1.0`

- [ ] Commit:
```bash
git add valt-cli/cmd/root.go valt-cli/main.go
git commit -m "feat(valt-cli): add Cobra root command with version"
```

---

## Success Criteria
- `go build ./...` compiles clean
- `./valt --version` prints version
- Config load/save round-trips without data loss
- API client returns structured errors on non-2xx responses
