package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sandwich-labs/invoice-generator-pro/internal/model"
	"github.com/sandwich-labs/invoice-generator-pro/internal/render"
	"github.com/sandwich-labs/invoice-generator-pro/internal/template"
)

var (
	inputFile    string
	outputFile   string
	templateName string
	outputFormat string
)

var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "Render an invoice from JSON data",
	Long: `Render an invoice from JSON data to PDF or HTML format.

The input JSON can be provided via a file path or piped through stdin.

Examples:
  # Render from file
  invgen render --input invoice.json --output invoice.pdf

  # Render from stdin
  cat invoice.json | invgen render --output invoice.pdf

  # Specify template and format
  invgen render --input invoice.json --template modern-blue --format html --output invoice.html`,
	RunE: runRender,
}

func init() {
	rootCmd.AddCommand(renderCmd)

	renderCmd.Flags().StringVarP(&inputFile, "input", "i", "", "input JSON file (or use stdin)")
	renderCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file path (required)")
	renderCmd.Flags().StringVarP(&templateName, "template", "t", "", "template name to use")
	renderCmd.Flags().StringVarP(&outputFormat, "format", "f", "", "output format: pdf or html")

	_ = renderCmd.MarkFlagRequired("output")
}

func runRender(cmd *cobra.Command, args []string) error {
	// Read input JSON
	var inputData []byte
	var err error

	if inputFile != "" {
		inputData, err = os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}
	} else {
		// Read from stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return fmt.Errorf("no input provided: use --input flag or pipe JSON to stdin")
		}
		inputData, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
	}

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
	if err := renderer.Render(inputData, tmplName, outputFile, format); err != nil {
		return fmt.Errorf("render failed: %w", err)
	}

	fmt.Printf("Generated %s: %s\n", format, outputFile)
	return nil
}
