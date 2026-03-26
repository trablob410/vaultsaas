package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var createWorkspaceCmd = &cobra.Command{
	Use:   "create-workspace",
	Short: "Create a new workspace within an organization",
	RunE:  runCreateWorkspace,
}

var (
	createWorkspaceOrgID string
	createWorkspaceName  string
	createWorkspaceSlug  string
)

func init() {
	createWorkspaceCmd.Flags().StringVar(&createWorkspaceOrgID, "org-id", "", "Organization ID (required)")
	createWorkspaceCmd.Flags().StringVar(&createWorkspaceName, "name", "", "Workspace name (required)")
	createWorkspaceCmd.Flags().StringVar(&createWorkspaceSlug, "slug", "", "Workspace slug (required)")
	createWorkspaceCmd.MarkFlagRequired("name") //nolint:errcheck
	createWorkspaceCmd.MarkFlagRequired("slug") //nolint:errcheck
	rootCmd.AddCommand(createWorkspaceCmd)
}

func runCreateWorkspace(cmd *cobra.Command, args []string) error {
	client := MustClient()

	orgID := createWorkspaceOrgID
	if orgID == "" {
		return fmt.Errorf("--org-id is required")
	}

	var resp struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := client.Post(context.Background(), "/orgs/"+orgID+"/workspaces", map[string]interface{}{
		"name": createWorkspaceName,
		"slug": createWorkspaceSlug,
	}, &resp); err != nil {
		return fmt.Errorf("creating workspace: %w", err)
	}

	fmt.Printf("Created workspace: %s (slug: %s, ID: %s)\n", resp.Name, resp.Slug, resp.ID)
	return nil
}
