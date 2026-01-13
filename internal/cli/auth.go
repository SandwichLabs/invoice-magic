package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sandwich-labs/invoice-generator-pro/internal/auth"
)

var (
	reauth     bool
	showScopes bool
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Google to access Sheets",
	Long: `Authenticate with Google to enable reading invoice data from Google Sheets.

This command initiates the OAuth2 flow to obtain permissions for reading
Google Sheets data. The authentication token is cached locally for future use.

Prerequisites:
  1. Create a Google Cloud project
  2. Enable the Google Sheets API
  3. Create OAuth2 credentials (Desktop app type)
  4. Download the credentials JSON file
  5. Set the path in config.yaml or use --credentials flag

Examples:
  # Authenticate with default credentials file from config
  invgen auth

  # Force re-authentication
  invgen auth --reauth

  # Show current authentication status
  invgen auth --show-scopes`,
	RunE: runAuth,
}

func init() {
	rootCmd.AddCommand(authCmd)

	authCmd.Flags().BoolVar(&reauth, "reauth", false, "force re-authentication")
	authCmd.Flags().BoolVar(&showScopes, "show-scopes", false, "show current token scopes")
}

func runAuth(cmd *cobra.Command, args []string) error {
	credentialsFile := viper.GetString("google.credentials_file")
	tokenFile := viper.GetString("google.token_file")

	if credentialsFile == "" {
		return fmt.Errorf("google.credentials_file not configured in config.yaml")
	}

	// Check if credentials file exists
	if _, err := os.Stat(credentialsFile); os.IsNotExist(err) {
		return fmt.Errorf("credentials file not found: %s\n\nTo set up Google Sheets access:\n"+
			"1. Go to https://console.cloud.google.com/\n"+
			"2. Create a project and enable the Google Sheets API\n"+
			"3. Create OAuth2 credentials (Desktop app)\n"+
			"4. Download and save to: %s", credentialsFile, credentialsFile)
	}

	authenticator, err := auth.NewAuthenticator(credentialsFile, tokenFile, auth.DefaultScopes())
	if err != nil {
		return fmt.Errorf("create authenticator: %w", err)
	}

	// Show current scopes if requested
	if showScopes {
		scopes, err := authenticator.GetGrantedScopes()
		if err != nil {
			fmt.Println("Not authenticated. Run 'invgen auth' to authenticate.")
			return nil
		}
		fmt.Println("Current authentication status: Authenticated")
		fmt.Println("Granted permissions:")
		for _, feature := range auth.ScopesToFeatures(scopes) {
			fmt.Printf("  - %s\n", feature)
		}
		return nil
	}

	// Clear token if re-auth requested
	if reauth {
		if err := authenticator.ClearToken(); err != nil {
			return fmt.Errorf("clear token: %w", err)
		}
		fmt.Println("Cleared existing token. Starting re-authentication...")
	}

	// Check if already authenticated
	if !authenticator.NeedsAuth() && !reauth {
		fmt.Println("Already authenticated with Google.")
		scopes, _ := authenticator.GetGrantedScopes()
		fmt.Println("Granted permissions:")
		for _, feature := range auth.ScopesToFeatures(scopes) {
			fmt.Printf("  - %s\n", feature)
		}
		fmt.Println("\nUse --reauth to force re-authentication.")
		return nil
	}

	// Initiate OAuth flow
	fmt.Println("Starting Google OAuth2 authentication...")
	fmt.Println("A browser window will open for you to grant access.")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	client, err := authenticator.GetClient(ctx)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	if client != nil {
		fmt.Println("\nAuthentication successful!")
		scopes, _ := authenticator.GetGrantedScopes()
		fmt.Println("Granted permissions:")
		for _, feature := range auth.ScopesToFeatures(scopes) {
			fmt.Printf("  - %s\n", feature)
		}
		fmt.Printf("\nToken saved to: %s\n", tokenFile)
		fmt.Println("\nYou can now use 'invgen render --sheets' to render invoices from Google Sheets.")
	}

	return nil
}
