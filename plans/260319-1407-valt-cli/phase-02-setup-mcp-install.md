# Phase 02: `valt setup` + `valt mcp install`

**Priority:** P0 — onboarding; without this, new users can't start
**Status:** COMPLETED — 2026-03-19

## Commands

```
valt setup
  1. Prompt for API URL (default: https://api.valt.dev)
  2. Open browser to {api_url}/api/v1/auth/google (OAuth login)
  3. Poll /api/v1/auth/cli-token until user completes login (short-lived token exchange)
  4. Store token in OS keychain
  5. Fetch /api/v1/orgs → list orgs → prompt user to pick one
  6. Fetch projects → prompt to pick project → save project_id to config
  7. Create an agent token via POST /api/v1/projects/{id}/agents + tokens
  8. Store agent token in keychain under "valt-agent-token"
  9. Print: ✓ Setup complete. Run: valt mcp install --ide claude

valt mcp install --ide claude|cursor|vscode
  → Prints or writes MCP server config block to the correct IDE config path
```

## Backend Prerequisite: CLI token exchange endpoint

The Go API needs a short-lived token endpoint for the CLI OAuth flow:
- `GET /api/v1/auth/cli-start` → returns `{session_id, login_url}` (creates pending session)
- `GET /api/v1/auth/cli-poll?session={id}` → returns `{status: "pending"|"complete", token?}`
- OAuth callback sets the session token when user completes login

**New migration needed:** `000030_cli_auth_sessions.up.sql`
- Table: `cli_auth_sessions (id UUID, token TEXT, expires_at TIMESTAMPTZ, created_at TIMESTAMPTZ)`

## Files

| File | Change |
|------|--------|
| `server/internal/database/migrations/000030_cli_auth_sessions.up.sql` | Create |
| `server/internal/auth/cli_session.go` | Create — session store + handlers |
| `server/cmd/server/main.go` | Modify — register CLI auth routes |
| `valt-cli/cmd/setup.go` | Create |
| `valt-cli/cmd/mcp.go` | Create |

---

## Task 1: Backend — CLI auth session

**Files:**
- Create: `server/internal/database/migrations/000030_cli_auth_sessions.up.sql`
- Create: `server/internal/auth/cli_session.go`

- [ ] Write migration:

```sql
CREATE TABLE cli_auth_sessions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token      TEXT,              -- set when login completes
    expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '5 minutes',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

- [ ] Write `cli_session.go` with two handlers:

```go
package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valt-dev/valt/server/pkg/apierror"
)

type CLISessionHandler struct {
	pool    *pgxpool.Pool
	jwtSvc  *Service   // existing JWT service
	baseURL string
}

func NewCLISessionHandler(pool *pgxpool.Pool, jwtSvc *Service, baseURL string) *CLISessionHandler {
	return &CLISessionHandler{pool: pool, jwtSvc: jwtSvc, baseURL: baseURL}
}

// Start handles GET /auth/cli-start
// Creates a pending session, returns session_id + login_url.
func (h *CLISessionHandler) Start(w http.ResponseWriter, r *http.Request) {
	var sessionID string
	err := h.pool.QueryRow(r.Context(),
		`INSERT INTO cli_auth_sessions DEFAULT VALUES RETURNING id`,
	).Scan(&sessionID)
	if err != nil {
		apierror.InternalError(w, "failed to create CLI session")
		return
	}
	loginURL := h.baseURL + "/api/v1/auth/google?cli_session=" + sessionID
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"session_id": sessionID,
		"login_url":  loginURL,
	})
}

// Poll handles GET /auth/cli-poll?session={id}
// Returns pending or complete+token.
func (h *CLISessionHandler) Poll(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session")
	var token *string
	var expiresAt time.Time
	err := h.pool.QueryRow(r.Context(),
		`SELECT token, expires_at FROM cli_auth_sessions WHERE id = $1`, sessionID,
	).Scan(&token, &expiresAt)
	if err != nil || time.Now().After(expiresAt) {
		apierror.NotFound(w, "session not found or expired")
		return
	}
	if token == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "pending"})
		return
	}
	// Delete session after successful poll
	_, _ = h.pool.Exec(r.Context(), `DELETE FROM cli_auth_sessions WHERE id = $1`, sessionID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "complete",
		"token":  *token,
	})
}
```

- [ ] Modify `auth/oauth.go` callback to set `cli_session` token if `cli_session` query param present:

```go
// In Google OAuth callback, after issuing JWT:
if cliSession := r.URL.Query().Get("cli_session"); cliSession != "" {
    _, _ = pool.Exec(r.Context(),
        `UPDATE cli_auth_sessions SET token = $1 WHERE id = $2 AND expires_at > NOW()`,
        accessToken, cliSession,
    )
}
```

- [ ] Register public routes:
```go
r.Get("/api/v1/auth/cli-start", cliSessionHandler.Start)
r.Get("/api/v1/auth/cli-poll",  cliSessionHandler.Poll)
```

- [ ] Run: `make test-unit`

- [ ] Commit:
```bash
git add server/internal/database/migrations/000030_* server/internal/auth/cli_session.go server/internal/auth/oauth.go server/cmd/server/main.go
git commit -m "feat(auth): add CLI OAuth token exchange endpoints for valt CLI setup"
```

---

## Task 2: valt-cli/cmd/setup.go

**Files:**
- Create: `valt-cli/cmd/setup.go`

- [ ] Write `setup` command:

```go
package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/valt-dev/valt/valt-cli/internal/api"
	"github.com/valt-dev/valt/valt-cli/internal/config"
	"github.com/valt-dev/valt/valt-cli/internal/keychain"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive setup wizard — configure API URL, login, pick project",
	RunE:  runSetup,
}

func init() { rootCmd.AddCommand(setupCmd) }

func runSetup(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	// Step 1: API URL
	fmt.Print("Valt API URL [https://api.valt.dev]: ")
	apiURL, _ := reader.ReadString('\n')
	apiURL = strings.TrimSpace(apiURL)
	if apiURL == "" { apiURL = "https://api.valt.dev" }

	client := api.New(apiURL, "")

	// Step 2: Start CLI session
	var session struct {
		SessionID string `json:"session_id"`
		LoginURL  string `json:"login_url"`
	}
	if err := client.Get(context.Background(), "/auth/cli-start", &session); err != nil {
		return fmt.Errorf("failed to start login session: %w", err)
	}

	fmt.Println("\nOpening browser for login...")
	openBrowser(session.LoginURL)
	fmt.Println("Login URL:", session.LoginURL)
	fmt.Println("Waiting for login...")

	// Step 3: Poll for token
	token, err := pollForToken(client, session.SessionID)
	if err != nil { return err }

	if err := keychain.SetToken(token); err != nil {
		return fmt.Errorf("storing token: %w", err)
	}
	fmt.Println("✓ Logged in")

	// Step 4: Pick project
	authedClient := api.New(apiURL, token)
	var orgsResp struct {
		Orgs []struct { ID string `json:"id"`; Name string `json:"name"` } `json:"orgs"`
	}
	if err := authedClient.Get(context.Background(), "/orgs", &orgsResp); err != nil {
		return fmt.Errorf("fetching orgs: %w", err)
	}
	if len(orgsResp.Orgs) == 0 {
		fmt.Println("No orgs found. Create one in the dashboard first.")
		return nil
	}
	fmt.Println("\nOrganizations:")
	for i, o := range orgsResp.Orgs { fmt.Printf("  %d. %s\n", i+1, o.Name) }
	fmt.Print("Pick org [1]: ")
	orgChoice, _ := reader.ReadString('\n')
	orgIdx := parseChoice(strings.TrimSpace(orgChoice), 1) - 1
	if orgIdx < 0 || orgIdx >= len(orgsResp.Orgs) { orgIdx = 0 }
	orgID := orgsResp.Orgs[orgIdx].ID

	// Fetch projects (simplified — first workspace)
	var wsResp struct {
		Workspaces []struct { ID string `json:"id"` } `json:"workspaces"`
	}
	authedClient.Get(context.Background(), "/orgs/"+orgID+"/workspaces", &wsResp)
	var projectID string
	if len(wsResp.Workspaces) > 0 {
		var projResp struct {
			Projects []struct { ID string `json:"id"`; Name string `json:"name"` } `json:"projects"`
		}
		authedClient.Get(context.Background(), "/workspaces/"+wsResp.Workspaces[0].ID+"/projects", &projResp)
		if len(projResp.Projects) > 0 {
			fmt.Println("\nProjects:")
			for i, p := range projResp.Projects { fmt.Printf("  %d. %s\n", i+1, p.Name) }
			fmt.Print("Pick project [1]: ")
			projChoice, _ := reader.ReadString('\n')
			projIdx := parseChoice(strings.TrimSpace(projChoice), 1) - 1
			if projIdx < 0 || projIdx >= len(projResp.Projects) { projIdx = 0 }
			projectID = projResp.Projects[projIdx].ID
		}
	}

	// Save config
	cfg := &config.Config{APIURL: apiURL, ProjectID: projectID}
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Println("\n✓ Setup complete!")
	fmt.Println("Next: valt mcp install --ide claude")
	return nil
}

func pollForToken(client *api.Client, sessionID string) (string, error) {
	deadline := time.Now().Add(3 * time.Minute)
	for time.Now().Before(deadline) {
		var result struct {
			Status string `json:"status"`
			Token  string `json:"token"`
		}
		if err := client.Get(context.Background(), "/auth/cli-poll?session="+sessionID, &result); err != nil {
			return "", err
		}
		if result.Status == "complete" && result.Token != "" {
			return result.Token, nil
		}
		time.Sleep(2 * time.Second)
	}
	return "", fmt.Errorf("login timed out after 3 minutes")
}

func openBrowser(url string) {
	var cmd string
	switch runtime.GOOS {
	case "windows": cmd = "start"
	case "darwin": cmd = "open"
	default: cmd = "xdg-open"
	}
	exec.Command(cmd, url).Start()
}

func parseChoice(s string, def int) int {
	n := def
	fmt.Sscan(s, &n)
	return n
}
```

- [ ] Register in `root.go` init (already done via `init()`)

- [ ] Run: `cd valt-cli && go build ./...`

- [ ] Commit:
```bash
git add valt-cli/cmd/setup.go
git commit -m "feat(valt-cli): add interactive setup wizard with OAuth browser flow"
```

---

## Task 3: valt-cli/cmd/mcp.go — `valt mcp install`

**Files:**
- Create: `valt-cli/cmd/mcp.go`

- [ ] Write `mcp install` command:

```go
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{Use: "mcp", Short: "MCP server management"}
var mcpInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Output or write MCP config for your IDE",
	RunE:  runMCPInstall,
}

var ideFlag string

func init() {
	mcpInstallCmd.Flags().StringVar(&ideFlag, "ide", "", "Target IDE: claude, cursor, vscode")
	mcpCmd.AddCommand(mcpInstallCmd)
	rootCmd.AddCommand(mcpCmd)
}

// mcpServerConfig is the JSON block to inject into IDE config.
func mcpServerConfig(binaryPath string) map[string]interface{} {
	return map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"valt": map[string]interface{}{
				"command": binaryPath,
				"args":    []string{},
			},
		},
	}
}

func runMCPInstall(cmd *cobra.Command, args []string) error {
	// Find valt-mcp-server binary path
	mcpBinary, err := findMCPBinary()
	if err != nil {
		fmt.Println("valt-mcp-server binary not found in PATH.")
		fmt.Println("Download from: https://github.com/valt-dev/valt/releases")
		return err
	}

	cfg := mcpServerConfig(mcpBinary)
	block, _ := json.MarshalIndent(cfg, "", "  ")

	switch ideFlag {
	case "claude":
		return writeIDEConfig(claudeConfigPath(), block, "Claude Desktop")
	case "cursor":
		return writeIDEConfig(cursorConfigPath(), block, "Cursor")
	default:
		fmt.Println("Add this to your IDE's MCP config:\n")
		fmt.Println(string(block))
	}
	return nil
}

func writeIDEConfig(path string, block []byte, ideName string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	// Read existing config, merge mcpServers key
	existing := map[string]interface{}{}
	if data, err := os.ReadFile(path); err == nil {
		json.Unmarshal(data, &existing)
	}
	var newCfg map[string]interface{}
	json.Unmarshal(block, &newCfg)
	if servers, ok := newCfg["mcpServers"]; ok {
		existing["mcpServers"] = servers
	}
	out, _ := json.MarshalIndent(existing, "", "  ")
	if err := os.WriteFile(path, out, 0644); err != nil {
		return fmt.Errorf("writing %s config: %w", ideName, err)
	}
	fmt.Printf("✓ %s config updated: %s\n", ideName, path)
	fmt.Println("Restart your IDE to load the MCP server.")
	return nil
}

func findMCPBinary() (string, error) {
	// Check PATH for valt-mcp-server
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		p := filepath.Join(dir, "valt-mcp-server")
		if runtime.GOOS == "windows" { p += ".exe" }
		if _, err := os.Stat(p); err == nil { return p, nil }
	}
	return "", fmt.Errorf("valt-mcp-server not in PATH")
}

func claudeConfigPath() string {
	home, _ := os.UserHomeDir()
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "Claude", "claude_desktop_config.json")
	}
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	}
	return filepath.Join(home, ".config", "claude", "claude_desktop_config.json")
}

func cursorConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cursor", "mcp.json")
}
```

- [ ] Run: `cd valt-cli && go build ./... && ./valt-cli mcp install`
  Expected: prints JSON config block

- [ ] Commit:
```bash
git add valt-cli/cmd/mcp.go
git commit -m "feat(valt-cli): add mcp install command for Claude/Cursor IDE config"
```

---

## Success Criteria
- `valt setup` opens browser, waits for OAuth, saves token + project to config
- `valt mcp install --ide claude` writes correct JSON to Claude Desktop config path
- `valt mcp install` (no flag) prints JSON block to stdout for manual paste
- Setup works on Mac, Linux, Windows
