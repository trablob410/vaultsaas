package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var createProjectCmd = &cobra.Command{
	Use:   "create-project",
	Short: "Create a new project within a workspace",
	RunE:  runCreateProject,
}

var (
	createProjectWorkspaceID string
	createProjectName        string
	createProjectSlug        string
)

func init() {
	createProjectCmd.Flags().StringVar(&createProjectWorkspaceID, "workspace-id", "", "Workspace ID (required)")
	createProjectCmd.Flags().StringVar(&createProjectName, "name", "", "Project name (required)")
	createProjectCmd.Flags().StringVar(&createProjectSlug, "slug", "", "Project slug (required)")
	createProjectCmd.MarkFlagRequired("name") //nolint:errcheck
	createProjectCmd.MarkFlagRequired("slug") //nolint:errcheck
	rootCmd.AddCommand(createProjectCmd)
}

func runCreateProject(cmd *cobra.Command, args []string) error {
	client := MustClient()

	workspaceID := createProjectWorkspaceID
	if workspaceID == "" {
		return fmt.Errorf("--workspace-id is required")
	}

	var resp struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := client.Post(context.Background(), "/workspaces/"+workspaceID+"/projects", map[string]interface{}{
		"name": createProjectName,
		"slug": createProjectSlug,
	}, &resp); err != nil {
		return fmt.Errorf("creating project: %w", err)
	}

	fmt.Printf("Created project: %s (slug: %s, ID: %s)\n", resp.Name, resp.Slug, resp.ID)
	return nil
}
