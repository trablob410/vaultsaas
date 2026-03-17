package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var runProjectID string

var runCmd = &cobra.Command{
	Use:   "run [-- cmd args...]",
	Short: "Run a command with secrets injected as environment variables",
	Long: `Fetch approved credentials for a project and inject them as env vars,
then execute the given command.

Example:
  valt run --project <id> -- node app.js`,
	DisableFlagParsing: false,
	RunE:               runRun,
}

func init() {
	runCmd.Flags().StringVar(&runProjectID, "project", "", "Project ID to fetch credentials for (required)")
	_ = runCmd.MarkFlagRequired("project")
}

func runRun(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified — use: valt run --project <id> -- <command> [args...]")
	}

	fmt.Printf("Fetching credentials for project %s...\n", runProjectID)

	// TODO(Phase 15): replace stub with real credential fetch
	// c, err := newClientFromConfig()
	// credentials, err := c.ListCredentials(runProjectID)
	// for _, cred := range credentials {
	//     secretEnv[strings.ToUpper(cred.Name)] = cred.Value
	// }

	secretEnv := map[string]string{} // stub: no secrets injected yet

	if verbose {
		fmt.Printf("Injecting %d secret(s) as environment variables\n", len(secretEnv))
	}

	command := args[0]
	cmdArgs := args[1:]

	c := exec.Command(command, cmdArgs...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin

	// Inherit current env and inject secrets
	env := os.Environ()
	for k, v := range secretEnv {
		env = append(env, strings.ToUpper(k)+"="+v)
	}
	c.Env = env

	if err := c.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
