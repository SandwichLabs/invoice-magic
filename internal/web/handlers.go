package web

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/sandwich-labs/invoice-generator-pro/internal/render"
	"github.com/sandwich-labs/invoice-generator-pro/internal/sheets"
	tmpl "github.com/sandwich-labs/invoice-generator-pro/internal/template"
)

// WebHandlerConfig holds dependencies for creating a WebHandler
type WebHandlerConfig struct {
	SheetsSource *sheets.Source
	TemplateMgr  *tmpl.Manager
	Renderer     *render.Renderer
	TemplateDir  string // HTML templates directory
	WriteEnabled bool   // Whether write operations are enabled
}

// WebHandler handles web requests with injected dependencies
type WebHandler struct {
	sheetsSource  *sheets.Source
	templateMgr   *tmpl.Manager
	renderer      *render.Renderer
	pageTemplates map[string]*template.Template
	tempDir       string
	writeEnabled  bool
}

// InvoiceRowView represents a row for display
type InvoiceRowView struct {
	RowNumber     int
	InvoiceNumber string
	CustomerName  string
	Date          string
	TotalGross    float64
	Currency      string
	IsValid       bool
	ValidationErr string
}

// InvoiceListData for the invoices page
type InvoiceListData struct {
	Title     string
	Invoices  []InvoiceRowView
	Error     string
	SheetInfo string
}

// PreviewData for the preview page
type PreviewData struct {
	Title        string
	RowNumber    int
	Invoice      *InvoiceRowView
	Templates    []tmpl.Info
	SelectedTmpl string
	PreviewHTML  template.HTML
	Error        string
}

// NewWebHandler creates a new WebHandler with parsed templates
func NewWebHandler(config WebHandlerConfig) (*WebHandler, error) {
	templateDir := config.TemplateDir
	if templateDir == "" {
		templateDir = "./web/templates"
	}

	// Create temp directory for rendered files
	tempDir, err := os.MkdirTemp("", "invoice-web-*")
	if err != nil {
		return nil, fmt.Errorf("create temp directory: %w", err)
	}

	h := &WebHandler{
		sheetsSource:  config.SheetsSource,
		templateMgr:   config.TemplateMgr,
		renderer:      config.Renderer,
		pageTemplates: make(map[string]*template.Template),
		tempDir:       tempDir,
		writeEnabled:  config.WriteEnabled,
	}

	// Parse templates
	funcMap := template.FuncMap{
		"formatMoney": func(amount float64, currency string) string {
			if currency == "" {
				currency = "USD"
			}
			return fmt.Sprintf("%s %.2f", currency, amount)
		},
	}

	// Parse layout template
	layoutPath := filepath.Join(templateDir, "layout.html")
	layoutContent, err := os.ReadFile(layoutPath)
	if err != nil {
		return nil, fmt.Errorf("read layout template: %w", err)
	}

	// Parse page templates
	pages := []string{"invoices.html", "preview.html", "provision.html"}
	for _, page := range pages {
		pagePath := filepath.Join(templateDir, page)
		pageContent, err := os.ReadFile(pagePath)
		if err != nil {
			return nil, fmt.Errorf("read %s template: %w", page, err)
		}

		// Parse partials
		partialsDir := filepath.Join(templateDir, "partials")
		var partialContent string
		if entries, err := os.ReadDir(partialsDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && filepath.Ext(entry.Name()) == ".html" {
					content, err := os.ReadFile(filepath.Join(partialsDir, entry.Name()))
					if err == nil {
						partialContent += string(content)
					}
				}
			}
		}

		// Combine layout + page + partials
		combined := string(layoutContent) + string(pageContent) + partialContent
		t, err := template.New(page).Funcs(funcMap).Parse(combined)
		if err != nil {
			return nil, fmt.Errorf("parse %s template: %w", page, err)
		}
		h.pageTemplates[page] = t
	}

	return h, nil
}

// Cleanup removes temporary files
func (h *WebHandler) Cleanup() {
	os.RemoveAll(h.tempDir)
}

// render renders a template with the given data
func (h *WebHandler) render(w http.ResponseWriter, templateName string, data interface{}) {
	t, ok := h.pageTemplates[templateName]
	if !ok {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Execute the "layout" template which includes "content"
	if err := t.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
	}
}

// IndexRedirect redirects to the invoices page
func (h *WebHandler) IndexRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/invoices", http.StatusSeeOther)
}

// InvoicesPage displays the list of invoices from the sheet
func (h *WebHandler) InvoicesPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Fetch all invoices from configured sheet
	invoices, err := h.sheetsSource.FetchInvoices(ctx)
	if err != nil {
		data := InvoiceListData{
			Title: "Invoices",
			Error: fmt.Sprintf("Failed to fetch invoices: %v", err),
		}
		h.render(w, "invoices.html", data)
		return
	}

	// Convert to view models with validation status
	var views []InvoiceRowView
	for _, inv := range invoices {
		view := InvoiceRowView{
			RowNumber:     inv.RowNumber,
			InvoiceNumber: inv.Invoice.Meta.InvoiceNumber,
			CustomerName:  inv.Invoice.Customer.Name,
			Date:          inv.Invoice.Meta.Date,
			TotalGross:    inv.Invoice.Totals.Gross,
			Currency:      inv.Invoice.Meta.Currency,
		}
		if err := inv.Invoice.Validate(); err != nil {
			view.IsValid = false
			view.ValidationErr = err.Error()
		} else {
			view.IsValid = true
		}
		views = append(views, view)
	}

	data := InvoiceListData{
		Title:    "Invoices",
		Invoices: views,
	}

	h.render(w, "invoices.html", data)
}

// RefreshInvoices is an HTMX handler that returns just the invoice list
func (h *WebHandler) RefreshInvoices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Fetch all invoices from configured sheet
	invoices, err := h.sheetsSource.FetchInvoices(ctx)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<tr><td colspan="6" class="error">Failed to fetch invoices: %v</td></tr>`, err)
		return
	}

	// Convert and render table rows
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	for _, inv := range invoices {
		isValid := inv.Invoice.Validate() == nil
		validClass := "valid"
		if !isValid {
			validClass = "invalid"
		}

		currency := inv.Invoice.Meta.Currency
		if currency == "" {
			currency = "USD"
		}

		fmt.Fprintf(w, `<tr class="%s" onclick="window.location='/invoices/%d'" style="cursor: pointer;">
			<td>%d</td>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
			<td>%s %.2f</td>
			<td>%s</td>
		</tr>`,
			validClass,
			inv.RowNumber,
			inv.RowNumber,
			template.HTMLEscapeString(inv.Invoice.Meta.InvoiceNumber),
			template.HTMLEscapeString(inv.Invoice.Customer.Name),
			template.HTMLEscapeString(inv.Invoice.Meta.Date),
			currency,
			inv.Invoice.Totals.Gross,
			func() string {
				if isValid {
					return "Valid"
				}
				return "Invalid"
			}(),
		)
	}
}

// InvoicePreview displays the preview page for a single invoice
func (h *WebHandler) InvoicePreview(w http.ResponseWriter, r *http.Request) {
	rowStr := chi.URLParam(r, "row")
	rowNum, err := strconv.Atoi(rowStr)
	if err != nil {
		http.Error(w, "Invalid row number", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Fetch the invoice
	invoiceRow, err := h.sheetsSource.FetchInvoice(ctx, rowNum)
	if err != nil {
		data := PreviewData{
			Title:     "Invoice Preview",
			RowNumber: rowNum,
			Error:     fmt.Sprintf("Failed to fetch invoice: %v", err),
		}
		h.render(w, "preview.html", data)
		return
	}

	// Get available templates
	templates, _ := h.templateMgr.List()

	// Create invoice view
	invoiceView := &InvoiceRowView{
		RowNumber:     invoiceRow.RowNumber,
		InvoiceNumber: invoiceRow.Invoice.Meta.InvoiceNumber,
		CustomerName:  invoiceRow.Invoice.Customer.Name,
		Date:          invoiceRow.Invoice.Meta.Date,
		TotalGross:    invoiceRow.Invoice.Totals.Gross,
		Currency:      invoiceRow.Invoice.Meta.Currency,
	}
	if err := invoiceRow.Invoice.Validate(); err != nil {
		invoiceView.IsValid = false
		invoiceView.ValidationErr = err.Error()
	} else {
		invoiceView.IsValid = true
	}

	// Default template
	selectedTmpl := "default"
	if len(templates) > 0 {
		selectedTmpl = templates[0].Name
	}

	data := PreviewData{
		Title:        fmt.Sprintf("Invoice %s", invoiceRow.Invoice.Meta.InvoiceNumber),
		RowNumber:    rowNum,
		Invoice:      invoiceView,
		Templates:    templates,
		SelectedTmpl: selectedTmpl,
	}

	h.render(w, "preview.html", data)
}

// RenderInvoice is an HTMX handler that renders the invoice with the selected template
func (h *WebHandler) RenderInvoice(w http.ResponseWriter, r *http.Request) {
	rowStr := chi.URLParam(r, "row")
	rowNum, err := strconv.Atoi(rowStr)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<div class="error">Invalid row number</div>`)
		return
	}

	if err := r.ParseForm(); err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<div class="error">Failed to parse form: %v</div>`, err)
		return
	}

	templateName := r.FormValue("template")
	if templateName == "" {
		templateName = "default"
	}

	ctx := r.Context()

	// Fetch the invoice
	invoiceRow, err := h.sheetsSource.FetchInvoice(ctx, rowNum)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<div class="error">Failed to fetch invoice: %v</div>`, err)
		return
	}

	// Render to temp HTML file
	tempPath := filepath.Join(h.tempDir, fmt.Sprintf("preview-%d.html", rowNum))
	err = h.renderer.Render(invoiceRow.RawJSON, templateName, tempPath, "html")
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<div class="error">Render failed: %v</div>`, err)
		return
	}

	// Read rendered HTML
	htmlBytes, err := os.ReadFile(tempPath)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<div class="error">Failed to read rendered file: %v</div>`, err)
		return
	}

	// Return the rendered HTML in an iframe container
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<iframe srcdoc="%s" class="preview-frame"></iframe>`,
		template.HTMLEscapeString(string(htmlBytes)))
}

// DownloadPDF generates and serves a PDF download
func (h *WebHandler) DownloadPDF(w http.ResponseWriter, r *http.Request) {
	rowStr := chi.URLParam(r, "row")
	rowNum, err := strconv.Atoi(rowStr)
	if err != nil {
		http.Error(w, "Invalid row number", http.StatusBadRequest)
		return
	}

	templateName := r.URL.Query().Get("template")
	if templateName == "" {
		templateName = "default"
	}

	ctx := r.Context()

	// Fetch the invoice
	invoiceRow, err := h.sheetsSource.FetchInvoice(ctx, rowNum)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch invoice: %v", err), http.StatusInternalServerError)
		return
	}

	// Generate PDF to temp file
	filename := fmt.Sprintf("invoice_%s.pdf", invoiceRow.Invoice.Meta.InvoiceNumber)
	pdfPath := filepath.Join(h.tempDir, filename)
	err = h.renderer.Render(invoiceRow.RawJSON, templateName, pdfPath, "pdf")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to render PDF: %v", err), http.StatusInternalServerError)
		return
	}

	// Serve file download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/pdf")
	http.ServeFile(w, r, pdfPath)
}

// ProvisionData for the provision page
type ProvisionData struct {
	Title        string
	Headers      []HeaderMapping
	WriteEnabled bool
	Success      bool
	Error        string
}

// HeaderMapping represents a column header mapping
type HeaderMapping struct {
	Column string
	Name   string
}

// ProvisionPage displays the provision page
func (h *WebHandler) ProvisionPage(w http.ResponseWriter, r *http.Request) {
	headers := h.sheetsSource.GetHeaderMapping()

	var mappings []HeaderMapping
	for i, name := range headers {
		mappings = append(mappings, HeaderMapping{
			Column: columnLetter(i),
			Name:   name,
		})
	}

	data := ProvisionData{
		Title:        "Provision Headers",
		Headers:      mappings,
		WriteEnabled: h.writeEnabled,
	}

	h.render(w, "provision.html", data)
}

// ProvisionHeaders handles the provision action
func (h *WebHandler) ProvisionHeaders(w http.ResponseWriter, r *http.Request) {
	if !h.writeEnabled {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<div class="error">Write operations not enabled. Start server with --write flag or use CLI: invgen provision</div>`)
		return
	}

	ctx := r.Context()

	if err := h.sheetsSource.ProvisionHeaders(ctx); err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<div class="error">Failed to provision headers: %v</div>`, err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<div class="success">Headers provisioned successfully!</div>`)
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
