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
