# Product Requirements Document: invoice-generator-pro (MVP)

## 1. Executive Summary
**invoice-generator-pro** is a high-performance, stateless CLI utility designed to transform structured JSON data into professional-grade PDF and HTML invoices. Built on Go and the Typst typesetting engine, it serves as a "rendering microservice" for other applications. This MVP focuses exclusively on the orchestration of document synthesis and template management, removing the need for internal database persistence in favor of real-time data ingestion.

---

## 2. Product Objectives
- **Typographic Excellence:** Leverage Typst to produce documents that surpass LaTeX and HTML-to-PDF tools in visual quality and rendering speed.
- **Portability:** Provide a single, statically linked Go binary that can be deployed in CI/CD pipelines, Docker containers, or local developer environments.
- **Developer-Centric:** Use standard JSON as the bridge between business logic and document layout.
- **Customizability:** Allow end-users to manage and modify templates without recompiling the application.

---

## 3. Functional Requirements

### 3.1 Core Rendering Engine
- **Input:** The system must accept invoice data in a structured JSON format via a file path or standard input (stdin).
- **Process:** The system will invoke the Typst CLI, passing the JSON data as a metadata dictionary using the `--input` flag.
- **Output:** The system must generate high-fidelity PDF (archival standard) and structural HTML (for email previews).
- **Concurrency:** The engine must support concurrent rendering requests if triggered by a batch process.

### 3.2 Template Management
- **Template Registry:** The system must maintain a directory (e.g., `./templates`) containing `.typ` files.
- **Discovery:** Users must be able to specify a template by name (e.g., `modern-blue`) via a CLI flag.
- **Scaffolding:** A command to initialize a new, compliant Typst template based on a "Base" standard.
- **Static Assets:** The system must correctly resolve paths for company logos and signatures stored within the template directory.

### 3.3 CLI Interface (Cobra-based)
- **`render` Command:** The primary entry point. 
    - Flags: `--template`, `--input`, `--output`, `--format` (pdf|html).
- **`template` Command:** 
    - `list`: Show available invoice layouts.
    - `init`: Create a boilerplate template.
- **Configuration:** Use Viper to define a default `template_dir` and `output_dir` in a `config.yaml` file.

---

## 4. Technical Specifications

### 4.1 Technology Stack
- **Language:** Go 1.21+ (utilizing `io/fs` and `os/exec`).
- **CLI Framework:** Cobra (Command definition) & Viper (Configuration).
- **Typesetting Engine:** Typst CLI (External dependency for MVP).
- **Data Format:** JSON (Standard library `encoding/json`).

### 4.2 Proposed Project Structure
```text
invoice-generator-pro/
├── cmd/
│   └── invgen/             # Main entry point
├── internal/
│   ├── cli/                # Cobra command definitions
│   ├── render/             # Typst orchestration logic
│   ├── template/           # Filesystem abstraction (io/fs)
│   └── model/              # Go structs for Invoice/LineItems (validation only)
├── templates/
│   ├── default.typ         # Standard invoice layout
│   └── assets/             # Default logos/fonts
└── config.yaml             # User settings
```

### 4.3 Input Data Contract (JSON)
The system expects a schema similar to the following:
```json
{
  "meta": { "invoice_number": "INV-2025-001", "date": "2025-01-10" },
  "sender": { "name": "ACME Corp", "address": "123 Go Lane", "tax_id": "DE12345678" },
  "customer": { "name": "John Doe", "email": "john@example.com" },
  "items": [
    { "description": "Software Consulting", "qty": 10, "unit_price": 150.00, "vat": 0.19 }
  ],
  "totals": { "net": 1500.00, "tax": 285.00, "gross": 1785.00 }
}
```

---

## 5. Non-Functional Requirements
- **Execution Speed:** Rendering a single-page PDF should take <200ms once initialized.
- **Security:** 
    - The Typst process must be restricted to the `template_dir` root.
    - JSON input must be sanitized to prevent Typst code injection.
- **Reliability:** Standard Error (stderr) from the Typst compiler must be captured and returned as a Go error for debugging.

---

## 6. User Stories

### 6.1 The Automation Engineer
*"As an engineer, I want to pipe a JSON stream from my billing database into `invgen` so that I can generate 500 monthly invoices in a single shell loop without setting up a database for the generator itself."*

### 6.2 The Graphic Designer
*"As a designer, I want to edit the `.typ` file in the `templates/` folder and see the changes reflected immediately in the PDF output without touching the Go source code."*

---

## 7. MVP Implementation Roadmap

### Phase 1: The Wrapper (Week 1)
- Set up Cobra/Viper CLI.
- Implement `render` command logic to call `typst compile`.
- Create the `internal/render` package to handle JSON-to-Typst input mapping.

### Phase 2: Template System (Week 2)
- Implement `internal/template` using `io/fs` to manage the template directory.
- Create a "Standard" Typst template that consumes the JSON data.
- Implement the `template list` command.

### Phase 3: Validation & Polish (Week 3)
- Add JSON schema validation to ensure the input data is complete before calling Typst.
- Implement HTML output support.
- Document CLI usage and template variables.

---

## 8. Success Metrics
- **Zero-State Success:** A user can download the binary and generate a sample invoice using the default template in under 60 seconds.
- **Correctness:** Mathematical totals in the PDF match the JSON input exactly (no floating-point drift).
- **Stability:** 100% success rate for rendering valid JSON inputs.

---

## 9. Future Scope (Post-MVP)
- **Embedded Templates:** Bundling default templates inside the binary using `//go:embed`.
- **S3/Cloud Support:** Directly uploading generated PDFs to AWS S3.
- **QR Code Generation:** Integration of GiroCode/EPC QR codes via Typst packages.
- **Webhook Integration:** Sending a POST request to a callback URL once rendering is complete.
