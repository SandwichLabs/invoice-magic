// testserve starts invgen's web UI backed by an in-memory repository seeded
// from testdata/sample_invoice.json. It requires no Google auth and is used
// for local development and browser-based QA.
//
// Usage: go run ./cmd/testserve [--port 9191]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/sandwich-labs/invoice-generator-pro/internal/model"
	"github.com/sandwich-labs/invoice-generator-pro/internal/repository/memrepo"
	"github.com/sandwich-labs/invoice-generator-pro/internal/template"
	"github.com/sandwich-labs/invoice-generator-pro/internal/web"
)

func main() {
	port := flag.Int("port", 9191, "port to listen on")
	flag.Parse()

	// Seed repository with sample invoices from testdata/
	repo := memrepo.New()
	fixtures := []string{
		"testdata/sample_invoice.json",
	}
	for _, path := range fixtures {
		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("warn: skipping %s: %v", path, err)
			continue
		}
		var inv model.Invoice
		if err := json.Unmarshal(data, &inv); err != nil {
			log.Fatalf("parse %s: %v", path, err)
		}
		if err := repo.Seed(&inv); err != nil {
			log.Fatalf("seed %s: %v", path, err)
		}
		log.Printf("seeded: %s", inv.Meta.InvoiceNumber)
	}

	// Add a second synthetic invoice so the list is interesting
	inv2 := model.Invoice{
		Meta: model.Meta{
			InvoiceNumber: "INV-2025-002",
			Date:          "2025-02-01",
			DueDate:       "2025-03-01",
			Currency:      "USD",
		},
		Sender: model.Party{Name: "ACME Corporation", Email: "billing@acme.example.com"},
		Customer: model.Party{
			Name:    "Jane Smith",
			Company: "Smith & Co",
			Email:   "jane@smith.example.com",
		},
		Items: []model.LineItem{
			{Description: "Annual Support Contract", Qty: 1, UnitPrice: 4800, VAT: 0},
		},
		Totals: model.Totals{Net: 4800, Tax: 0, Gross: 4800},
		Notes:  "Renewed annually. Thank you!",
	}
	if err := repo.Seed(&inv2); err != nil {
		log.Fatalf("seed inv2: %v", err)
	}

	tmplMgr := template.NewManager("./templates")

	cfg := web.ServerConfig{
		Port:        *port,
		Repo:        repo,
		SheetsSource: nil, // no sheets backend; provision page shows unavailable message
		TemplateMgr: tmplMgr,
		TemplateDir: "./web/templates",
		StaticDir:   "./web/static",
	}

	fmt.Printf("Test server running at http://localhost:%d\n", *port)
	fmt.Printf("Backend: in-memory (2 invoices seeded)\n")
	if err := web.StartServer(cfg); err != nil {
		log.Fatal(err)
	}
}
