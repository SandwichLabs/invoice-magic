package render

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sandwich-labs/invoice-generator-pro/internal/template"
)

// Renderer handles invoice rendering via Typst
type Renderer struct {
	templateMgr *template.Manager
}

// New creates a new Renderer with the given template manager
func New(tmplMgr *template.Manager) *Renderer {
	return &Renderer{
		templateMgr: tmplMgr,
	}
}

// Render renders an invoice to the specified output file
func (r *Renderer) Render(jsonData []byte, templateName, outputPath, format string) error {
	// Get template path
	tmplPath := r.templateMgr.GetPath(templateName)
	if tmplPath == "" {
		return fmt.Errorf("template not found: %s", templateName)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if outputDir != "" && outputDir != "." {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Escape JSON for shell
	escapedJSON := escapeJSON(string(jsonData))

	// Build typst command
	args := []string{
		"compile",
		"--input", fmt.Sprintf("data=%s", escapedJSON),
	}

	// Add format-specific options
	if format == "html" {
		args = append(args, "--format", "html")
	}

	// Add input and output paths
	args = append(args, tmplPath, outputPath)

	// Execute typst
	cmd := exec.Command("typst", args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := stderr.String()
		if stderrStr != "" {
			return fmt.Errorf("typst compilation failed: %s", stderrStr)
		}
		return fmt.Errorf("typst compilation failed: %w", err)
	}

	return nil
}

// escapeJSON escapes a JSON string for safe use as a Typst input parameter
func escapeJSON(s string) string {
	// Replace backslashes first, then quotes
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

// CheckTypstInstalled verifies that the Typst CLI is available
func CheckTypstInstalled() error {
	cmd := exec.Command("typst", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("typst CLI not found in PATH: please install from https://github.com/typst/typst")
	}
	return nil
}
