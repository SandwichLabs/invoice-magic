package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sandwich-labs/invoice-generator-pro/internal/auth"
	"github.com/sandwich-labs/invoice-generator-pro/internal/model"
	"github.com/sandwich-labs/invoice-generator-pro/internal/render"
	"github.com/sandwich-labs/invoice-generator-pro/internal/sheets"
	"github.com/sandwich-labs/invoice-generator-pro/internal/template"
)

var (
	inputFile     string
	outputFile    string
	templateName  string
	outputFormat  string
	useSheets     bool
	sheetsRow     int
	spreadsheetID string
	sheetName     string
)

var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "Render an invoice from JSON data or Google Sheets",
	Long: `Render an invoice from JSON data or Google Sheets to PDF or HTML format.

The input can be provided via:
  - A JSON file path (--input)
  - Piped through stdin
  - Google Sheets (--sheets flag)

Examples:
  # Render from file
  invgen render --input invoice.json --output invoice.pdf

  # Render from stdin
  cat invoice.json | invgen render --output invoice.pdf

  # Render from Google Sheets (all invoices)
  invgen render --sheets --output ./invoices/

  # Render specific row from Google Sheets
  invgen render --sheets --row 5 --output invoice.pdf

  # Render from specific spreadsheet
  invgen render --sheets --spreadsheet-id "abc123" --sheet-name "Invoices" --output ./invoices/

  # Specify template and format
  invgen render --input invoice.json --template modern-blue --format html --output invoice.html`,
	RunE: runRender,
}

func init() {
	rootCmd.AddCommand(renderCmd)

	renderCmd.Flags().StringVarP(&inputFile, "input", "i", "", "input JSON file (or use stdin)")
	renderCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file or directory path (required)")
	renderCmd.Flags().StringVarP(&templateName, "template", "t", "", "template name to use")
	renderCmd.Flags().StringVarP(&outputFormat, "format", "f", "", "output format: pdf or html")

	// Google Sheets flags
	renderCmd.Flags().BoolVar(&useSheets, "sheets", false, "read invoice data from Google Sheets")
	renderCmd.Flags().IntVar(&sheetsRow, "row", 0, "specific row to render (0 = all rows)")
	renderCmd.Flags().StringVar(&spreadsheetID, "spreadsheet-id", "", "Google Spreadsheet ID (overrides config)")
	renderCmd.Flags().StringVar(&sheetName, "sheet-name", "", "Sheet name/tab (overrides config)")

	_ = renderCmd.MarkFlagRequired("output")
}

func runRender(cmd *cobra.Command, args []string) error {
	// Handle Google Sheets source
	if useSheets {
		return runRenderFromSheets()
	}

	// Read input JSON from file or stdin
	var inputData []byte
	var err error

	if inputFile != "" {
		inputData, err = os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}
	} else {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return fmt.Errorf("no input provided: use --input flag, --sheets flag, or pipe JSON to stdin")
		}
		inputData, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
	}

	return renderInvoiceData(inputData, outputFile)
}

func runRenderFromSheets() error {
	// Get Google credentials
	credentialsFile := viper.GetString("google.credentials_file")
	tokenFile := viper.GetString("google.token_file")

	if credentialsFile == "" {
		return fmt.Errorf("google.credentials_file not configured. Run 'invgen auth' first")
	}

	// Create authenticator
	authenticator, err := auth.NewAuthenticator(credentialsFile, tokenFile, auth.DefaultScopes())
	if err != nil {
		return fmt.Errorf("create authenticator: %w", err)
	}

	if authenticator.NeedsAuth() {
		return fmt.Errorf("not authenticated with Google. Run 'invgen auth' first")
	}

	// Get HTTP client
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	httpClient, err := authenticator.GetClient(ctx)
	if err != nil {
		return fmt.Errorf("get authenticated client: %w", err)
	}

	// Create Sheets client
	sheetsClient, err := sheets.NewClient(ctx, httpClient)
	if err != nil {
		return fmt.Errorf("create sheets client: %w", err)
	}

	// Build source config from flags and viper config
	sourceConfig := buildSourceConfig()

	if sourceConfig.SpreadsheetID == "" {
		return fmt.Errorf("spreadsheet_id not configured. Set in config.yaml or use --spreadsheet-id flag")
	}

	// Create sheets source
	source := sheets.NewSource(sheetsClient, sourceConfig)

	// Fetch and render invoices
	if sheetsRow > 0 {
		// Render single invoice
		invoiceRow, err := source.FetchInvoice(ctx, sheetsRow)
		if err != nil {
			return fmt.Errorf("fetch invoice from row %d: %w", sheetsRow, err)
		}

		if err := invoiceRow.Invoice.Validate(); err != nil {
			return fmt.Errorf("invoice validation failed (row %d): %w", sheetsRow, err)
		}

		return renderInvoiceData(invoiceRow.RawJSON, outputFile)
	}

	// Render all invoices
	invoices, err := source.FetchInvoices(ctx)
	if err != nil {
		return fmt.Errorf("fetch invoices: %w", err)
	}

	if len(invoices) == 0 {
		fmt.Println("No invoices found in the sheet")
		return nil
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputFile, 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	format := outputFormat
	if format == "" {
		format = viper.GetString("default_format")
	}

	fmt.Printf("Rendering %d invoices from Google Sheets...\n", len(invoices))

	for _, inv := range invoices {
		if err := inv.Invoice.Validate(); err != nil {
			fmt.Printf("  Skipping row %d: %v\n", inv.RowNumber, err)
			continue
		}

		// Generate output filename
		filename := fmt.Sprintf("invoice_%s.%s", inv.Invoice.Meta.InvoiceNumber, format)
		outPath := filepath.Join(outputFile, filename)

		if err := renderInvoiceData(inv.RawJSON, outPath); err != nil {
			fmt.Printf("  Error rendering row %d: %v\n", inv.RowNumber, err)
			continue
		}

		fmt.Printf("  Generated: %s (row %d)\n", filename, inv.RowNumber)
	}

	return nil
}

func buildSourceConfig() sheets.SourceConfig {
	config := sheets.SourceConfig{
		SpreadsheetID: viper.GetString("sheets.spreadsheet_id"),
		SheetName:     viper.GetString("sheets.sheet_name"),
		HeaderRow:     viper.GetInt("sheets.header_row"),
		DataStartRow:  viper.GetInt("sheets.data_start_row"),
		Fields: sheets.FieldMappings{
			InvoiceNumber:   viper.GetString("sheets.fields.invoice_number"),
			Date:            viper.GetString("sheets.fields.date"),
			DueDate:         viper.GetString("sheets.fields.due_date"),
			Currency:        viper.GetString("sheets.fields.currency"),
			SenderName:      viper.GetString("sheets.fields.sender_name"),
			SenderCompany:   viper.GetString("sheets.fields.sender_company"),
			SenderAddress:   viper.GetString("sheets.fields.sender_address"),
			SenderTaxID:     viper.GetString("sheets.fields.sender_tax_id"),
			SenderEmail:     viper.GetString("sheets.fields.sender_email"),
			SenderPhone:     viper.GetString("sheets.fields.sender_phone"),
			CustomerName:    viper.GetString("sheets.fields.customer_name"),
			CustomerCompany: viper.GetString("sheets.fields.customer_company"),
			CustomerAddress: viper.GetString("sheets.fields.customer_address"),
			CustomerTaxID:   viper.GetString("sheets.fields.customer_tax_id"),
			CustomerEmail:   viper.GetString("sheets.fields.customer_email"),
			CustomerPhone:   viper.GetString("sheets.fields.customer_phone"),
			ItemDescription: viper.GetString("sheets.fields.item_description"),
			ItemQty:         viper.GetString("sheets.fields.item_qty"),
			ItemUnitPrice:   viper.GetString("sheets.fields.item_unit_price"),
			ItemVAT:         viper.GetString("sheets.fields.item_vat"),
			TotalNet:        viper.GetString("sheets.fields.total_net"),
			TotalTax:        viper.GetString("sheets.fields.total_tax"),
			TotalGross:      viper.GetString("sheets.fields.total_gross"),
			Notes:           viper.GetString("sheets.fields.notes"),
		},
	}

	// Override with command-line flags
	if spreadsheetID != "" {
		config.SpreadsheetID = spreadsheetID
	}
	if sheetName != "" {
		config.SheetName = sheetName
	}

	// Set defaults
	if config.SheetName == "" {
		config.SheetName = "Invoices"
	}
	if config.HeaderRow == 0 {
		config.HeaderRow = 1
	}
	if config.DataStartRow == 0 {
		config.DataStartRow = 2
	}

	return config
}

func renderInvoiceData(inputData []byte, outPath string) error {
	// Parse and validate invoice
	invoice, err := model.ParseInvoice(inputData)
	if err != nil {
		return fmt.Errorf("invalid invoice data: %w", err)
	}

	if err := invoice.Validate(); err != nil {
		return fmt.Errorf("invoice validation failed: %w", err)
	}

	// Resolve template
	tmplDir := viper.GetString("template_dir")
	tmplMgr := template.NewManager(tmplDir)

	tmplName := templateName
	if tmplName == "" {
		tmplName = viper.GetString("default_template")
	}

	if !tmplMgr.Exists(tmplName) {
		return fmt.Errorf("template not found: %s", tmplName)
	}

	// Resolve format
	format := outputFormat
	if format == "" {
		format = viper.GetString("default_format")
	}

	if format != "pdf" && format != "html" {
		return fmt.Errorf("invalid format: %s (must be pdf or html)", format)
	}

	// Render invoice
	renderer := render.New(tmplMgr)
	if err := renderer.Render(inputData, tmplName, outPath, format); err != nil {
		return fmt.Errorf("render failed: %w", err)
	}

	fmt.Printf("Generated %s: %s\n", format, outPath)
	return nil
}
