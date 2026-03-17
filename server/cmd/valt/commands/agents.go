package commands

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Manage AI agents",
}

var (
	agentProjectID string
	agentName      string
)

var agentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List agents for a project",
	RunE:  runAgentsList,
}

var agentsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new agent",
	RunE:  runAgentsCreate,
}

func init() {
	agentsListCmd.Flags().StringVar(&agentProjectID, "project", "", "Project ID (required)")
	_ = agentsListCmd.MarkFlagRequired("project")

	agentsCreateCmd.Flags().StringVar(&agentProjectID, "project", "", "Project ID (required)")
	agentsCreateCmd.Flags().StringVar(&agentName, "name", "", "Agent name (required)")
	_ = agentsCreateCmd.MarkFlagRequired("project")
	_ = agentsCreateCmd.MarkFlagRequired("name")

	agentsCmd.AddCommand(agentsListCmd, agentsCreateCmd)
}

func runAgentsList(cmd *cobra.Command, args []string) error {
	c, err := newClientFromConfig()
	if err != nil {
		return err
	}
	agents, err := c.ListAgents(agentProjectID)
	if err != nil {
		return err
	}
	if len(agents) == 0 {
		fmt.Println("No agents found.")
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tSTATUS")
	for _, a := range agents {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			strField(a, "id"),
			strField(a, "name"),
			strField(a, "status"),
		)
	}
	return w.Flush()
}

func runAgentsCreate(cmd *cobra.Command, args []string) error {
	c, err := newClientFromConfig()
	if err != nil {
		return err
	}
	result, err := c.CreateAgent(agentProjectID, agentName)
	if err != nil {
		return err
	}
	fmt.Printf("Agent created: %s\n", strField(result, "id"))
	if tok, ok := result["token"].(string); ok && tok != "" {
		fmt.Printf("Agent token: %s\n", tok)
		fmt.Println("Store this token securely — it will not be shown again.")
	}
	return nil
}
