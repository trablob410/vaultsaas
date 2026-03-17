package commands

import (
	"github.com/spf13/cobra"
)

var (
	apiURL  string
	token   string
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "valt",
	Short: "Valt secret vault CLI",
	Long:  "Manage secrets, agents, and approvals in your Valt vault.",
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "http://localhost:8080/api/v1", "Valt API URL")
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "Authentication token (or set VALT_TOKEN env var)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	rootCmd.AddCommand(loginCmd, secretsCmd, agentsCmd, scanCmd, runCmd)
}
