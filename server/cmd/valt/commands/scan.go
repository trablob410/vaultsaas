package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	scanPath    string
	scanUseMCP  bool
)

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan for hardcoded secrets",
	Long:  "Scan a directory for hardcoded secrets using the MCP scan_secrets tool.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runScan,
}

func init() {
	scanCmd.Flags().StringVar(&scanPath, "path", ".", "Directory to scan")
	scanCmd.Flags().BoolVar(&scanUseMCP, "mcp", false, "Trigger scan via HTTP MCP gateway")
}

func runScan(cmd *cobra.Command, args []string) error {
	targetPath := scanPath
	if len(args) > 0 {
		targetPath = args[0]
	}

	if scanUseMCP {
		fmt.Printf("Triggering MCP scan for path: %s\n", targetPath)
		fmt.Println("Use the MCP tool 'scan_secrets' via the HTTP MCP gateway.")
		fmt.Println("Example: POST /mcp with tool=scan_secrets, args={\"path\": \"" + targetPath + "\"}")
		return nil
	}

	fmt.Println("valt scan: Use the MCP tool 'scan_secrets' for local file scanning.")
	fmt.Printf("  Target path: %s\n", targetPath)
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  valt scan --mcp --path <path>   Trigger via HTTP MCP gateway")
	fmt.Println("  Use the MCP tool 'scan_secrets' directly in your MCP-native workflow")
	return nil
}
