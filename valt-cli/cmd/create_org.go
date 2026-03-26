package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var createOrgCmd = &cobra.Command{
	Use:   "create-org",
	Short: "Create a new organization",
	RunE:  runCreateOrg,
}

var (
	createOrgName string
	createOrgSlug string
)

func init() {
	createOrgCmd.Flags().StringVar(&createOrgName, "name", "", "Organization name (required)")
	createOrgCmd.Flags().StringVar(&createOrgSlug, "slug", "", "Organization slug (required)")
	createOrgCmd.MarkFlagRequired("name") //nolint:errcheck
	createOrgCmd.MarkFlagRequired("slug") //nolint:errcheck
	rootCmd.AddCommand(createOrgCmd)
}

func runCreateOrg(cmd *cobra.Command, args []string) error {
	client := MustClient()

	var resp struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := client.Post(context.Background(), "/orgs", map[string]interface{}{
		"name": createOrgName,
		"slug": createOrgSlug,
	}, &resp); err != nil {
		return fmt.Errorf("creating org: %w", err)
	}

	fmt.Printf("Created org: %s (slug: %s, ID: %s)\n", resp.Name, resp.Slug, resp.ID)
	return nil
}
