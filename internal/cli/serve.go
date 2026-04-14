package cli

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sandwich-labs/invoice-generator-pro/internal/auth"
	"github.com/sandwich-labs/invoice-generator-pro/internal/repository/sheetsrepo"
	"github.com/sandwich-labs/invoice-generator-pro/internal/sheets"
	"github.com/sandwich-labs/invoice-generator-pro/internal/template"
	"github.com/sandwich-labs/invoice-generator-pro/internal/web"
)

var (
	servePort   int
	writeEnable bool
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web server for invoice preview",
	Long: `Start a local web server that provides a browser interface
for viewing invoices from Google Sheets and previewing rendered output.

Requires prior authentication via 'invgen auth'.

Examples:
  # Start server on default port 8080
  invgen serve

  # Start server on custom port
  invgen serve --port 3000

  # Enable write operations (for provisioning headers)
  invgen serve --write`,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8080, "Port to run server on")
	serveCmd.Flags().BoolVar(&writeEnable, "write", false, "Enable write operations (requires read/write auth)")
}

func runServe(cmd *cobra.Command, args []string) error {
	// Check authentication
	credentialsFile := viper.GetString("google.credentials_file")
	tokenFile := viper.GetString("google.token_file")

	if credentialsFile == "" {
		return fmt.Errorf("google.credentials_file not configured. Set up config.yaml first")
	}

	// Select scopes based on write mode
	scopes := auth.DefaultScopes()
	if writeEnable {
		scopes = auth.AllScopes() // Read/write access
	}

	authenticator, err := auth.NewAuthenticator(credentialsFile, tokenFile, scopes)
	if err != nil {
		return fmt.Errorf("auth setup failed: %w", err)
	}

	if authenticator.NeedsAuth() {
		return fmt.Errorf("not authenticated with Google. Run 'invgen auth' first")
	}

	// Get authenticated HTTP client
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var httpClient *http.Client
	if writeEnable {
		// Use scope upgrade in case user has read-only token
		httpClient, err = authenticator.GetClientWithScopeUpgrade(ctx)
	} else {
		httpClient, err = authenticator.GetClient(ctx)
	}
	if err != nil {
		return fmt.Errorf("get authenticated client: %w", err)
	}

	// Build sheets source config
	sourceConfig := buildSourceConfig()

	if sourceConfig.SpreadsheetID == "" {
		return fmt.Errorf("sheets.spreadsheet_id not configured. Set it in config.yaml")
	}

	// Create sheets client
	sheetsClient, err := sheets.NewClient(ctx, httpClient)
	if err != nil {
		return fmt.Errorf("create sheets client: %w", err)
	}

	// Create repository
	repo := sheetsrepo.New(sheetsClient, sourceConfig, writeEnable)

	// Create template manager
	tmplDir := viper.GetString("template_dir")
	tmplMgr := template.NewManager(tmplDir)

	// Start server
	config := web.ServerConfig{
		Port:         servePort,
		Repo:         repo,
		SheetsSource: repo.Source(), // provision operations
		TemplateMgr:  tmplMgr,
		TemplateDir:  "./web/templates",
		StaticDir:    "./web/static",
	}

	fmt.Printf("Connecting to Google Sheets...\n")
	fmt.Printf("  Spreadsheet: %s\n", sourceConfig.SpreadsheetID)
	fmt.Printf("  Sheet: %s\n", sourceConfig.SheetName)
	if writeEnable {
		fmt.Printf("  Mode: Read/Write (provisioning enabled)\n")
	} else {
		fmt.Printf("  Mode: Read-only\n")
	}
	fmt.Println()

	return web.StartServer(config)
}
