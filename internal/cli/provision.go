package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sandwich-labs/invoice-generator-pro/internal/auth"
	"github.com/sandwich-labs/invoice-generator-pro/internal/sheets"
)

var provisionCmd = &cobra.Command{
	Use:   "provision",
	Short: "Provision field mappings to the remote Google Sheet",
	Long: `Push the configured field mappings as column headers to the remote Google Sheet.

This command writes the header row to the configured sheet based on your
field mappings in config.yaml. This saves the manual step of setting up
column headers in Google Sheets.

Note: This requires read/write access to Google Sheets. If you previously
authenticated with read-only access, you'll be prompted to re-authenticate.

Examples:
  # Push headers to the configured sheet
  invgen provision

  # Preview what would be written without making changes
  invgen provision --dry-run`,
	RunE: runProvision,
}

var dryRun bool

func init() {
	rootCmd.AddCommand(provisionCmd)
	provisionCmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview headers without writing to sheet")
}

func runProvision(cmd *cobra.Command, args []string) error {
	// Get config
	credentialsFile := viper.GetString("google.credentials_file")
	tokenFile := viper.GetString("google.token_file")

	if credentialsFile == "" {
		return fmt.Errorf("google.credentials_file not configured. Set up config.yaml first")
	}

	// Build source config to get field mappings
	sourceConfig := buildSourceConfig()

	if sourceConfig.SpreadsheetID == "" {
		return fmt.Errorf("sheets.spreadsheet_id not configured. Set it in config.yaml")
	}

	// Create a temporary source just to get the header mapping
	tempSource := sheets.NewSource(nil, sourceConfig)
	headers := tempSource.GetHeaderMapping()

	if len(headers) == 0 {
		return fmt.Errorf("no field mappings configured in config.yaml")
	}

	// Display what will be written
	fmt.Println("Field mappings to provision:")
	fmt.Println()
	for i, h := range headers {
		fmt.Printf("  Column %s: %s\n", columnLetter(i), h)
	}
	fmt.Println()
	fmt.Printf("Target: %s!A%d\n", sourceConfig.SheetName, sourceConfig.HeaderRow)
	fmt.Printf("Spreadsheet: %s\n", sourceConfig.SpreadsheetID)
	fmt.Println()

	if dryRun {
		fmt.Println("(Dry run - no changes made)")
		return nil
	}

	// Need read/write scope for provisioning
	authenticator, err := auth.NewAuthenticator(credentialsFile, tokenFile, auth.AllScopes())
	if err != nil {
		return fmt.Errorf("auth setup failed: %w", err)
	}

	// Get authenticated HTTP client (with scope upgrade if needed)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	httpClient, err := authenticator.GetClientWithScopeUpgrade(ctx)
	if err != nil {
		return fmt.Errorf("get authenticated client: %w", err)
	}

	// Create sheets client
	sheetsClient, err := sheets.NewClient(ctx, httpClient)
	if err != nil {
		return fmt.Errorf("create sheets client: %w", err)
	}

	// Create source with the sheets client
	source := sheets.NewSource(sheetsClient, sourceConfig)

	// Provision the headers
	fmt.Println("Writing headers to sheet...")
	if err := source.ProvisionHeaders(ctx); err != nil {
		return fmt.Errorf("provision headers: %w", err)
	}

	fmt.Println("Headers provisioned successfully!")
	return nil
}

// columnLetter converts a 0-indexed column number to a letter (A, B, ..., Z, AA, AB, ...).
func columnLetter(index int) string {
	result := ""
	for index >= 0 {
		result = string(rune('A'+index%26)) + result
		index = index/26 - 1
	}
	return result
}
