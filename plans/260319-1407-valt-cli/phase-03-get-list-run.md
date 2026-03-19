# Phase 03: `valt get` + `valt list` + `valt run`

**Priority:** P1 — daily developer workflow commands
**Status:** COMPLETED — 2026-03-19

## Commands

```bash
valt list [--project <id>]
  → GET /api/v1/secrets?project_id=<id>
  → prints table: NAME  TYPE  CREATED

valt get <secret-name>
  → GET /api/v1/secrets?name=<secret-name>&project_id=<id>
  → GET /api/v1/credentials/<request_id>  (if approved)
  → prints value to stdout (safe to pipe)

valt run -- <command> [args...]
  → fetches all accessible secrets in project
  → injects as env vars
  → exec.Command(command, args...) with extended env
```

## Files

| File | Change |
|------|--------|
| `valt-cli/cmd/list.go` | Create |
| `valt-cli/cmd/get.go` | Create |
| `valt-cli/cmd/run.go` | Create |

---

## Task 1: cmd/list.go

**Files:**
- Create: `valt-cli/cmd/list.go`

- [ ] Write:

```go
package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/valt-dev/valt/valt-cli/internal/config"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List accessible secrets in your project",
	RunE:  runList,
}

func init() { rootCmd.AddCommand(listCmd) }

func runList(cmd *cobra.Command, args []string) error {
	cfg, _ := config.Load()
	client := MustClient()

	path := "/secrets"
	if cfg.ProjectID != "" {
		path += "?project_id=" + cfg.ProjectID
	}

	var resp struct {
		Secrets []struct {
			ID             string `json:"id"`
			Name           string `json:"name"`
			CredentialType string `json:"credential_type"`
			CreatedAt      string `json:"created_at"`
		} `json:"secrets"`
		Total int `json:"total"`
	}
	if err := client.Get(context.Background(), path, &resp); err != nil {
		return fmt.Errorf("fetching secrets: %w", err)
	}

	if len(resp.Secrets) == 0 {
		fmt.Println("No secrets found.")
		return nil
	}

	fmt.Printf("%-30s %-20s %s\n", "NAME", "TYPE", "CREATED")
	fmt.Printf("%-30s %-20s %s\n", "----", "----", "-------")
	for _, s := range resp.Secrets {
		created := ""
		if len(s.CreatedAt) >= 10 { created = s.CreatedAt[:10] }
		fmt.Printf("%-30s %-20s %s\n", s.Name, s.CredentialType, created)
	}
	fmt.Printf("\nTotal: %d\n", resp.Total)
	return nil
}
```

- [ ] Run: `cd valt-cli && go build ./...`

- [ ] Commit:
```bash
git add valt-cli/cmd/list.go
git commit -m "feat(valt-cli): add valt list command"
```

---

## Task 2: cmd/get.go

**Files:**
- Create: `valt-cli/cmd/get.go`

- [ ] Write:

```go
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/valt-dev/valt/valt-cli/internal/config"
)

var getCmd = &cobra.Command{
	Use:   "get <secret-name>",
	Short: "Get a secret value (prints to stdout, safe to pipe)",
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

func init() { rootCmd.AddCommand(getCmd) }

func runGet(cmd *cobra.Command, args []string) error {
	name := args[0]
	cfg, _ := config.Load()
	client := MustClient()

	// Find secret by name
	path := "/secrets?name=" + name
	if cfg.ProjectID != "" { path += "&project_id=" + cfg.ProjectID }

	var secretsResp struct {
		Secrets []struct {
			ID             string `json:"id"`
			CredentialType string `json:"credential_type"`
		} `json:"secrets"`
	}
	if err := client.Get(context.Background(), path, &secretsResp); err != nil {
		return fmt.Errorf("searching secrets: %w", err)
	}
	if len(secretsResp.Secrets) == 0 {
		fmt.Fprintf(os.Stderr, "Secret '%s' not found\n", name)
		os.Exit(1)
	}
	secretID := secretsResp.Secrets[0].ID

	// Try to get active credential first
	var cred struct {
		Value  string `json:"value"`
		Status string `json:"status"`
	}
	err := client.Get(context.Background(), "/credentials/by-secret/"+secretID, &cred)
	if err == nil && cred.Value != "" {
		fmt.Print(cred.Value) // no newline — safe for $(valt get NAME)
		return nil
	}

	// No active credential — create access request
	fmt.Fprintf(os.Stderr, "No active credential for '%s'. Requesting access...\n", name)
	var req struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := client.Post(context.Background(),
		"/secrets/"+secretID+"/access-requests",
		map[string]interface{}{
			"reason":           "valt CLI get",
			"duration_minutes": 60,
		},
		&req,
	); err != nil {
		return fmt.Errorf("creating access request: %w", err)
	}

	if req.Status == "approved" {
		// Auto-approved (Tier 1)
		var c struct{ Value string `json:"value"` }
		client.Get(context.Background(), "/credentials/"+req.ID, &c)
		fmt.Print(c.Value)
		return nil
	}

	fmt.Fprintf(os.Stderr, "Access request created (ID: %s). Run: valt status %s\n", req.ID, req.ID)
	fmt.Fprintf(os.Stderr, "Waiting for approval. Check your email/Slack for the approval link.\n")
	return nil
}
```

- [ ] Run: `cd valt-cli && go build ./...`

- [ ] Commit:
```bash
git add valt-cli/cmd/get.go
git commit -m "feat(valt-cli): add valt get command with auto-request fallback"
```

---

## Task 3: cmd/run.go — inject all secrets as env vars

**Files:**
- Create: `valt-cli/cmd/run.go`

- [ ] Write the killer feature:

```go
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/valt-dev/valt/valt-cli/internal/config"
)

var runCmd = &cobra.Command{
	Use:                "run -- <command> [args...]",
	Short:              "Run a command with all accessible secrets injected as env vars",
	Args:               cobra.MinimumNArgs(1),
	RunE:               runRun,
	DisableFlagParsing: true, // pass all flags through to child command
}

func init() { rootCmd.AddCommand(runCmd) }

func runRun(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("command required: valt run -- <command>")
	}
	// Strip leading "--" separator if present
	if args[0] == "--" { args = args[1:] }
	if len(args) == 0 {
		return fmt.Errorf("command required after --")
	}

	cfg, _ := config.Load()
	client := MustClient()

	// Fetch all accessible secrets with active credentials
	path := "/credentials/active"
	if cfg.ProjectID != "" { path += "?project_id=" + cfg.ProjectID }

	var credsResp struct {
		Credentials []struct {
			SecretName string `json:"secret_name"`
			Value      string `json:"value"`
		} `json:"credentials"`
	}
	if err := client.Get(context.Background(), path, &credsResp); err != nil {
		return fmt.Errorf("fetching credentials: %w", err)
	}

	// Build env: start with current env, overlay secrets
	env := os.Environ()
	injected := 0
	for _, c := range credsResp.Credentials {
		if c.Value == "" { continue }
		envKey := strings.ToUpper(strings.ReplaceAll(c.SecretName, "-", "_"))
		env = append(env, envKey+"="+c.Value)
		injected++
	}

	fmt.Fprintf(os.Stderr, "valt: injecting %d secret(s)\n", injected)

	child := exec.Command(args[0], args[1:]...)
	child.Env = env
	child.Stdin = os.Stdin
	child.Stdout = os.Stdout
	child.Stderr = os.Stderr

	if err := child.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}
	return nil
}
```

**Backend prerequisite:** Add `GET /api/v1/credentials/active?project_id=<id>` endpoint that returns all active credential sessions with decrypted values for the authenticated user/agent.

- [ ] Run: `cd valt-cli && go build ./...`

- [ ] Run: `./valt-cli run -- env | grep VALT` (manual test after backend endpoint exists)

- [ ] Commit:
```bash
git add valt-cli/cmd/run.go
git commit -m "feat(valt-cli): add valt run — injects active secrets as env vars before exec"
```

---

## Success Criteria
- `valt list` prints table of secrets with name/type/date
- `valt get DB_PASSWORD` prints raw value (no newline) — works with `$(valt get X)`
- `valt get NONEXISTENT` exits non-zero with clear error
- `valt run -- node server.js` injects secrets as uppercase env vars and passes exit code through
- `valt run -- env | grep INJECTED_SECRET` shows the injected value
