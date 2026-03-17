package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var loginToken string

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with a Valt server",
	Long:  "Store your Valt authentication token locally.",
	RunE:  runLogin,
}

func init() {
	loginCmd.Flags().StringVar(&loginToken, "token", "", "Authentication token")
}

func runLogin(cmd *cobra.Command, args []string) error {
	tok := loginToken
	if tok == "" {
		// Interactive prompt
		fmt.Print("Enter your Valt token: ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading token: %w", err)
		}
		tok = strings.TrimSpace(input)
	}
	if tok == "" {
		return fmt.Errorf("token is required")
	}

	url := apiURL
	cfg, err := LoadConfig()
	if err == nil && cfg.APIURL != "" {
		url = cfg.APIURL
	}
	// --api-url flag overrides config
	if cmd.Root().PersistentFlags().Changed("api-url") {
		url = apiURL
	}

	if err := SaveConfig(&Config{Token: tok, APIURL: url}); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("Logged in successfully. Token stored at %s\n", configPath())
	return nil
}
