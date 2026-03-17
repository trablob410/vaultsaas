package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/valt-dev/valt/server/cmd/valt/client"
)

var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Manage secrets",
}

var secretsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List secrets",
	RunE:  runSecretsList,
}

var (
	secretName string
	secretType string
)

var secretsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new secret (prompts for value)",
	RunE:  runSecretsCreate,
}

func init() {
	secretsCreateCmd.Flags().StringVar(&secretName, "name", "", "Secret name (required)")
	secretsCreateCmd.Flags().StringVar(&secretType, "type", "generic", "Credential type")
	_ = secretsCreateCmd.MarkFlagRequired("name")
	secretsCmd.AddCommand(secretsListCmd, secretsCreateCmd)
}

func newClientFromConfig() (*client.Client, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	tok := token
	if tok == "" {
		tok = cfg.Token
	}
	url := apiURL
	if url == "http://localhost:8080/api/v1" && cfg.APIURL != "" {
		url = cfg.APIURL
	}
	return client.New(url, tok), nil
}

func runSecretsList(cmd *cobra.Command, args []string) error {
	c, err := newClientFromConfig()
	if err != nil {
		return err
	}
	secrets, err := c.ListSecrets()
	if err != nil {
		return err
	}
	if len(secrets) == 0 {
		fmt.Println("No secrets found.")
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tTYPE\tSTATUS")
	for _, s := range secrets {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			strField(s, "id"),
			strField(s, "name"),
			strField(s, "credential_type"),
			strField(s, "status"),
		)
	}
	return w.Flush()
}

func runSecretsCreate(cmd *cobra.Command, args []string) error {
	fmt.Print("Enter secret value: ")
	reader := bufio.NewReader(os.Stdin)
	value, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("reading value: %w", err)
	}
	value = strings.TrimSpace(value)

	c, err := newClientFromConfig()
	if err != nil {
		return err
	}
	result, err := c.CreateSecret(secretName, secretType, value)
	if err != nil {
		return err
	}
	fmt.Printf("Secret created: %s\n", strField(result, "id"))
	return nil
}

func strField(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return "-"
}
