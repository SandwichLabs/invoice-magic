package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	templateDir string
	outputDir   string

	// Version info set from main
	appVersion = "dev"
	appCommit  = "none"
	appDate    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "invgen",
	Short: "Generate professional invoices from JSON data",
	Long: `Invoice Generator Pro (invgen) transforms structured JSON data into
professional-grade PDF and HTML invoices using the Typst typesetting engine.

Examples:
  # Generate a PDF invoice
  invgen render --input invoice.json --output invoice.pdf

  # Generate HTML output
  invgen render --input invoice.json --output invoice.html --format html

  # Use a specific template
  invgen render --input invoice.json --template modern-blue

  # List available templates
  invgen template list`,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().StringVar(&templateDir, "template-dir", "", "directory containing templates")
	rootCmd.PersistentFlags().StringVar(&outputDir, "output-dir", "", "directory for output files")

	viper.BindPFlag("template_dir", rootCmd.PersistentFlags().Lookup("template-dir"))
	viper.BindPFlag("output_dir", rootCmd.PersistentFlags().Lookup("output-dir"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	}

	// Set defaults
	viper.SetDefault("template_dir", "./templates")
	viper.SetDefault("output_dir", "./output")
	viper.SetDefault("default_template", "default")
	viper.SetDefault("default_format", "pdf")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
		}
	}
}

// SetVersion sets the version information for the CLI
func SetVersion(version, commit, date string) {
	appVersion = version
	appCommit = commit
	appDate = date

	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}
