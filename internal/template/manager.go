package template

import (
	"os"
	"path/filepath"
	"strings"
)

// Info contains information about a template
type Info struct {
	Name string
	Path string
}

// Manager handles template discovery and management
type Manager struct {
	templateDir string
}

// NewManager creates a new template manager for the given directory
func NewManager(templateDir string) *Manager {
	return &Manager{
		templateDir: templateDir,
	}
}

// List returns all available templates
func (m *Manager) List() ([]Info, error) {
	entries, err := os.ReadDir(m.templateDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var templates []Info
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".typ") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".typ")
		templates = append(templates, Info{
			Name: name,
			Path: filepath.Join(m.templateDir, entry.Name()),
		})
	}

	return templates, nil
}

// Exists checks if a template with the given name exists
func (m *Manager) Exists(name string) bool {
	path := m.GetPath(name)
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

// GetPath returns the full path to a template file
func (m *Manager) GetPath(name string) string {
	// Check for exact .typ file
	path := filepath.Join(m.templateDir, name+".typ")
	if _, err := os.Stat(path); err == nil {
		return path
	}

	// Check if name already includes extension
	if strings.HasSuffix(name, ".typ") {
		path = filepath.Join(m.templateDir, name)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// DefaultTemplateContent returns the content for a new default template
func DefaultTemplateContent() string {
	return `// Invoice Template for invoice-generator-pro
// Edit this file to customize your invoice layout

// Parse invoice data from input
#let data = json(sys.inputs.data)

// Page setup
#set page(
  paper: "a4",
  margin: (top: 2cm, bottom: 2cm, left: 2cm, right: 2cm),
)

#set text(font: "Linux Libertine", size: 11pt)

// Header
#align(center)[
  #text(size: 24pt, weight: "bold")[INVOICE]
]

#v(1cm)

// Invoice metadata
#grid(
  columns: (1fr, 1fr),
  align: (left, right),
  [
    *Invoice Number:* #data.meta.invoice_number \
    *Date:* #data.meta.date \
    #if "due_date" in data.meta [*Due Date:* #data.meta.due_date]
  ],
  []
)

#v(1cm)

// Sender and Customer
#grid(
  columns: (1fr, 1fr),
  column-gutter: 2cm,
  [
    *From:* \
    #text(weight: "bold")[#data.sender.name] \
    #data.sender.address \
    #if "email" in data.sender [#data.sender.email] \
    #if "phone" in data.sender [#data.sender.phone]
  ],
  [
    *To:* \
    #text(weight: "bold")[#data.customer.name] \
    #if "company" in data.customer [#data.customer.company \]
    #if "address" in data.customer [#data.customer.address] \
    #if "email" in data.customer [#data.customer.email]
  ]
)

#v(1cm)

// Line items table
#table(
  columns: (auto, 1fr, auto, auto, auto),
  align: (center, left, right, right, right),
  stroke: 0.5pt,
  inset: 8pt,

  // Header
  [*#*], [*Description*], [*Qty*], [*Unit Price*], [*Amount*],

  // Items
  ..data.items.enumerate().map(((i, item)) => (
    [#(i + 1)],
    [#item.description],
    [#item.qty],
    [#item.unit_price],
    [#calc.round(item.qty * item.unit_price, digits: 2)],
  )).flatten()
)

#v(0.5cm)

// Totals
#align(right)[
  #table(
    columns: (auto, auto),
    align: (left, right),
    stroke: none,
    inset: 4pt,

    [*Subtotal:*], [#data.totals.net],
    [*Tax:*], [#data.totals.tax],
    table.hline(stroke: 1pt),
    [*Total:*], [*#data.totals.gross*],
  )
]

#v(1cm)

// Notes
#if "notes" in data [
  #line(length: 100%, stroke: 0.5pt)
  #v(0.3cm)
  *Notes:* #data.notes
]
`
}
