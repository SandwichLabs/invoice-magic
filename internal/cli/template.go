package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sandwich-labs/invoice-generator-pro/internal/template"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage invoice templates",
	Long:  `Commands for listing and initializing invoice templates.`,
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available templates",
	Long:  `List all available invoice templates in the template directory.`,
	RunE:  runTemplateList,
}

var templateInitCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Initialize a new template",
	Long: `Initialize a new invoice template based on the default template.

Example:
  invgen template init my-custom-template`,
	Args: cobra.ExactArgs(1),
	RunE: runTemplateInit,
}

func init() {
	rootCmd.AddCommand(templateCmd)
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateInitCmd)
}

func runTemplateList(cmd *cobra.Command, args []string) error {
	tmplDir := viper.GetString("template_dir")
	mgr := template.NewManager(tmplDir)

	templates, err := mgr.List()
	if err != nil {
		return fmt.Errorf("failed to list templates: %w", err)
	}

	if len(templates) == 0 {
		fmt.Println("No templates found in", tmplDir)
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tPATH")
	fmt.Fprintln(w, "----\t----")
	for _, t := range templates {
		fmt.Fprintf(w, "%s\t%s\n", t.Name, t.Path)
	}
	w.Flush()

	return nil
}

func runTemplateInit(cmd *cobra.Command, args []string) error {
	name := args[0]
	tmplDir := viper.GetString("template_dir")

	// Check if template already exists
	mgr := template.NewManager(tmplDir)
	if mgr.Exists(name) {
		return fmt.Errorf("template already exists: %s", name)
	}

	// Create new template file
	newPath := filepath.Join(tmplDir, name+".typ")

	// Copy default template content
	defaultContent := template.DefaultTemplateContent()

	if err := os.WriteFile(newPath, []byte(defaultContent), 0644); err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	fmt.Printf("Created new template: %s\n", newPath)
	fmt.Println("Edit the .typ file to customize your invoice layout.")

	return nil
}
