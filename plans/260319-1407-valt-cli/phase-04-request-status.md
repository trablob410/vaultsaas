# Phase 04: `valt request` + `valt status`

**Priority:** P2
**Status:** COMPLETED — 2026-03-19

## Commands

```bash
valt request <secret-name> --reason "fixing prod bug" --duration 2h
  → POST /api/v1/secrets/{id}/access-requests
  → prints: Request ID: <id>  Status: pending
  → prints: Check status: valt status <id>

valt status <request-id>
  → GET /api/v1/access-requests/{id}
  → prints status table
  → if approved: prints: valt get <secret-name> to retrieve value
```

## Files

| File | Change |
|------|--------|
| `valt-cli/cmd/request.go` | Create |
| `valt-cli/cmd/status.go` | Create |

---

## Task 1: cmd/request.go

**Files:**
- Create: `valt-cli/cmd/request.go`

- [ ] Write:

```go
package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/valt-dev/valt/valt-cli/internal/config"
)

var requestCmd = &cobra.Command{
	Use:   "request <secret-name>",
	Short: "Create an access request for a secret",
	Args:  cobra.ExactArgs(1),
	RunE:  runRequest,
}

var (
	requestReason   string
	requestDuration string
)

func init() {
	requestCmd.Flags().StringVar(&requestReason, "reason", "", "Reason for access (required for Tier 2+)")
	requestCmd.Flags().StringVar(&requestDuration, "duration", "1h", "Access duration (e.g. 30m, 2h, 8h)")
	rootCmd.AddCommand(requestCmd)
}

func runRequest(cmd *cobra.Command, args []string) error {
	name := args[0]
	cfg, _ := config.Load()
	client := MustClient()

	// Find secret by name
	path := "/secrets?name=" + name
	if cfg.ProjectID != "" { path += "&project_id=" + cfg.ProjectID }
	var secretsResp struct {
		Secrets []struct{ ID string `json:"id"` } `json:"secrets"`
	}
	if err := client.Get(context.Background(), path, &secretsResp); err != nil {
		return fmt.Errorf("searching for secret: %w", err)
	}
	if len(secretsResp.Secrets) == 0 {
		return fmt.Errorf("secret '%s' not found", name)
	}

	durationMin := parseDuration(requestDuration)

	var req struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := client.Post(context.Background(),
		"/secrets/"+secretsResp.Secrets[0].ID+"/access-requests",
		map[string]interface{}{
			"reason":           requestReason,
			"duration_minutes": durationMin,
		},
		&req,
	); err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	fmt.Printf("Request ID: %s\nStatus:     %s\n", req.ID, req.Status)
	if req.Status == "pending" {
		fmt.Printf("\nCheck status:  valt status %s\n", req.ID)
		fmt.Println("Approvers have been notified via email/Slack.")
	} else if req.Status == "approved" {
		fmt.Printf("\nAuto-approved. Get value:  valt get %s\n", name)
	}
	return nil
}

// parseDuration converts "2h", "30m", "1h30m" to minutes.
func parseDuration(s string) int {
	total := 0
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else if c == 'h' {
			total += n * 60; n = 0
		} else if c == 'm' {
			total += n; n = 0
		}
	}
	if total == 0 { total = 60 } // default 1h
	return total
}
```

- [ ] Run: `cd valt-cli && go build ./...`

- [ ] Commit:
```bash
git add valt-cli/cmd/request.go
git commit -m "feat(valt-cli): add valt request command to create access requests"
```

---

## Task 2: cmd/status.go

**Files:**
- Create: `valt-cli/cmd/status.go`

- [ ] Write:

```go
package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status <request-id>",
	Short: "Check approval status of an access request",
	Args:  cobra.ExactArgs(1),
	RunE:  runStatus,
}

func init() { rootCmd.AddCommand(statusCmd) }

func runStatus(cmd *cobra.Command, args []string) error {
	requestID := args[0]
	client := MustClient()

	var req struct {
		ID         string  `json:"id"`
		SecretName string  `json:"secret_name"`
		Status     string  `json:"status"`
		Reason     *string `json:"reason"`
		CreatedAt  string  `json:"created_at"`
		ExpiresAt  *string `json:"expires_at"`
	}
	if err := client.Get(context.Background(), "/access-requests/"+requestID, &req); err != nil {
		return fmt.Errorf("fetching request: %w", err)
	}

	fmt.Printf("Request:  %s\n", req.ID)
	fmt.Printf("Secret:   %s\n", req.SecretName)
	fmt.Printf("Status:   %s\n", req.Status)
	if req.Reason != nil { fmt.Printf("Reason:   %s\n", *req.Reason) }
	if len(req.CreatedAt) >= 19 { fmt.Printf("Created:  %s\n", req.CreatedAt[:19]) }
	if req.ExpiresAt != nil && len(*req.ExpiresAt) >= 19 {
		fmt.Printf("Expires:  %s\n", (*req.ExpiresAt)[:19])
	}

	switch req.Status {
	case "approved":
		fmt.Printf("\n✓ Approved. Retrieve:  valt get %s\n", req.SecretName)
	case "pending":
		fmt.Println("\n⏳ Pending approval. Approver notified via email/Slack.")
	case "rejected":
		fmt.Println("\n✗ Rejected. Create a new request with a more detailed reason.")
	case "expired":
		fmt.Println("\n⚠ Expired. Create a new request.")
	}
	return nil
}
```

- [ ] Run: `cd valt-cli && go build ./...`

- [ ] Commit:
```bash
git add valt-cli/cmd/status.go
git commit -m "feat(valt-cli): add valt status command to check request approval"
```

---

## Success Criteria
- `valt request db-password --reason "fixing prod" --duration 2h` creates request and prints ID
- `valt status <id>` shows status with human-readable next step
- Duration parsing: `2h` → 120, `30m` → 30, `1h30m` → 90
