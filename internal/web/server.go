package web

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/sandwich-labs/invoice-generator-pro/internal/render"
	"github.com/sandwich-labs/invoice-generator-pro/internal/sheets"
	"github.com/sandwich-labs/invoice-generator-pro/internal/template"
)

// ServerConfig holds configuration for the web server
type ServerConfig struct {
	Port         int
	SheetsClient *sheets.Client
	SourceConfig sheets.SourceConfig
	TemplateMgr  *template.Manager
	TemplateDir  string // HTML templates directory
	StaticDir    string // Static files directory
	WriteEnabled bool   // Whether write operations are enabled
}

// StartServer initializes and starts the HTTP server
func StartServer(config ServerConfig) error {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Create handler with dependencies
	sheetsSource := sheets.NewSource(config.SheetsClient, config.SourceConfig)
	renderer := render.New(config.TemplateMgr)

	handler, err := NewWebHandler(WebHandlerConfig{
		SheetsSource: sheetsSource,
		TemplateMgr:  config.TemplateMgr,
		Renderer:     renderer,
		TemplateDir:  config.TemplateDir,
		WriteEnabled: config.WriteEnabled,
	})
	if err != nil {
		return fmt.Errorf("create web handler: %w", err)
	}

	// Static files
	staticDir := config.StaticDir
	if staticDir == "" {
		staticDir = "./web/static"
	}
	fs := http.FileServer(http.Dir(staticDir))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	// Routes
	r.Get("/", handler.IndexRedirect)
	r.Get("/invoices", handler.InvoicesPage)
	r.Post("/invoices/refresh", handler.RefreshInvoices)
	r.Get("/invoices/{row}", handler.InvoicePreview)
	r.Post("/invoices/{row}/render", handler.RenderInvoice)
	r.Get("/invoices/{row}/download", handler.DownloadPDF)

	// Provision routes
	r.Get("/provision", handler.ProvisionPage)
	r.Post("/provision", handler.ProvisionHeaders)

	addr := fmt.Sprintf(":%d", config.Port)
	log.Printf("Starting invoice server on http://localhost%s", addr)
	return http.ListenAndServe(addr, r)
}
