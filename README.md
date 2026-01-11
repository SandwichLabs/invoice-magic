# invoice-generator-pro

A high-performance CLI utility that transforms structured JSON data into professional-grade PDF and HTML invoices using Go and the [Typst](https://typst.app) typesetting engine.

## Features

- **Fast rendering** - Generate invoices in under 200ms
- **Multiple formats** - Output to PDF or HTML
- **Template system** - Customizable Typst templates
- **Stdin support** - Pipe JSON data directly from other tools
- **Validation** - Input validation before rendering
- **Cross-platform** - Builds for Linux, macOS, and Windows

## Requirements

- [Typst CLI](https://github.com/typst/typst) must be installed and available in PATH

```bash
# Install Typst (macOS/Linux via Homebrew)
brew install typst

# Or download from GitHub releases
# https://github.com/typst/typst/releases
```

## Installation

### From Source

```bash
git clone https://github.com/sandwich-labs/invoice-generator-pro.git
cd invoice-generator-pro
go build -o invgen ./cmd/invgen
```

### Using Task

```bash
task build
```

## Usage

### Generate a PDF Invoice

```bash
invgen render --input invoice.json --output invoice.pdf
```

### Generate HTML Output

```bash
invgen render --input invoice.json --output invoice.html --format html
```

### Use a Specific Template

```bash
invgen render --input invoice.json --output invoice.pdf --template modern-blue
```

### Pipe from Stdin

```bash
cat invoice.json | invgen render --output invoice.pdf

# Or from another command
curl -s https://api.example.com/invoice/123 | invgen render --output invoice.pdf
```

### List Available Templates

```bash
invgen template list
```

### Create a New Template

```bash
invgen template init my-custom-template
```

## Input JSON Schema

The input JSON must follow this structure:

```json
{
  "meta": {
    "invoice_number": "INV-2025-001",
    "date": "2025-01-10",
    "due_date": "2025-02-10"
  },
  "sender": {
    "name": "ACME Corp",
    "address": "123 Business Lane\nNew York, NY 10001",
    "email": "billing@acme.com",
    "phone": "+1 555-123-4567",
    "tax_id": "US12345678"
  },
  "customer": {
    "name": "John Doe",
    "company": "Doe Industries",
    "address": "456 Client Street\nLos Angeles, CA 90001",
    "email": "john@doe.com"
  },
  "items": [
    {
      "description": "Software Consulting",
      "qty": 10,
      "unit_price": 150.00,
      "vat": 0.19
    },
    {
      "description": "Project Management",
      "qty": 5,
      "unit_price": 100.00,
      "vat": 0.19
    }
  ],
  "totals": {
    "net": 2000.00,
    "tax": 380.00,
    "gross": 2380.00
  },
  "notes": "Payment due within 30 days. Thank you for your business!"
}
```

### Required Fields

| Field | Description |
|-------|-------------|
| `meta.invoice_number` | Unique invoice identifier |
| `meta.date` | Invoice date |
| `sender.name` | Sender/company name |
| `customer.name` | Customer name |
| `items` | Array of line items (at least one) |
| `items[].description` | Item description |
| `items[].qty` | Quantity (positive number) |
| `items[].unit_price` | Price per unit (non-negative) |
| `totals.net` | Subtotal before tax |
| `totals.tax` | Tax amount |
| `totals.gross` | Total including tax |

## Configuration

Create a `config.yaml` in your working directory:

```yaml
# Directory containing Typst templates
template_dir: ./templates

# Directory for generated output files
output_dir: ./output

# Default template to use if not specified
default_template: default

# Default output format (pdf or html)
default_format: pdf
```

Configuration can also be set via CLI flags:

```bash
invgen render --template-dir /path/to/templates --input invoice.json --output invoice.pdf
```

## Creating Custom Templates

Templates are [Typst](https://typst.app/docs) files (`.typ`) that receive invoice data via `sys.inputs.data`.

### Basic Template Structure

```typst
// Parse invoice data from input
#let data = json(sys.inputs.data)

// Page setup
#set page(paper: "a4", margin: 2cm)
#set text(size: 11pt)

// Header
#align(center)[
  #text(size: 24pt, weight: "bold")[INVOICE]
]

// Use data fields
*Invoice Number:* #data.meta.invoice_number
*Date:* #data.meta.date

// Iterate over items
#for item in data.items [
  #item.description: #item.qty x #item.unit_price
]

// Totals
*Total:* #data.totals.gross
```

### Initialize from Default

```bash
invgen template init my-template
# Creates templates/my-template.typ with the default layout
```

## Development

### Prerequisites

- Go 1.21+
- [Task](https://taskfile.dev) (optional, for build automation)
- [golangci-lint](https://golangci-lint.run) (for linting)

### Build Commands

```bash
task build          # Build the binary
task test           # Run tests
task lint           # Run linter
task clean          # Remove build artifacts
task run            # Build and run
```

### Project Structure

```
invoice-generator-pro/
├── cmd/invgen/          # CLI entry point
├── internal/
│   ├── cli/             # Cobra command definitions
│   ├── model/           # Invoice data structures
│   ├── render/          # Typst orchestration
│   └── template/        # Template management
├── templates/           # Default templates
├── testdata/            # Test fixtures
├── config.yaml          # Default configuration
└── Taskfile.yml         # Build automation
```

### Running Tests

```bash
go test -v ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## License

Apache-2.0

## Acknowledgments

- [Typst](https://typst.app) - The modern typesetting system powering invoice rendering
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
