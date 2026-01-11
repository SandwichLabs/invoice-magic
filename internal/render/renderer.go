package render

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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

	// Write JSON to a temporary file (Typst reads input from file paths)
	tmpFile, err := os.CreateTemp("", "invoice-*.json")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err := tmpFile.Write(jsonData); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	_ = tmpFile.Close()

	// Get absolute path for the temp file
	absPath, err := filepath.Abs(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Build typst command
	// Use --root / to allow absolute paths in json() function
	args := []string{
		"compile",
		"--root", "/",
		"--input", fmt.Sprintf("data=%s", absPath),
	}

	// Add format-specific options
	if format == "html" {
		// HTML export requires experimental features flag
		args = append(args, "--features", "html", "--format", "html")
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

// CheckTypstInstalled verifies that the Typst CLI is available
func CheckTypstInstalled() error {
	cmd := exec.Command("typst", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("typst CLI not found in PATH: please install from https://github.com/typst/typst")
	}
	return nil
}
