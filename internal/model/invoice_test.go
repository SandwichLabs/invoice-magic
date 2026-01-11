package model

import (
	"testing"
)

func TestParseInvoice(t *testing.T) {
	validJSON := `{
		"meta": {"invoice_number": "INV-001", "date": "2025-01-10"},
		"sender": {"name": "ACME Corp"},
		"customer": {"name": "John Doe"},
		"items": [{"description": "Service", "qty": 1, "unit_price": 100, "vat": 0.1}],
		"totals": {"net": 100, "tax": 10, "gross": 110}
	}`

	invoice, err := ParseInvoice([]byte(validJSON))
	if err != nil {
		t.Fatalf("ParseInvoice failed: %v", err)
	}

	if invoice.Meta.InvoiceNumber != "INV-001" {
		t.Errorf("expected invoice_number INV-001, got %s", invoice.Meta.InvoiceNumber)
	}

	if invoice.Sender.Name != "ACME Corp" {
		t.Errorf("expected sender name ACME Corp, got %s", invoice.Sender.Name)
	}

	if len(invoice.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(invoice.Items))
	}
}

func TestParseInvoice_InvalidJSON(t *testing.T) {
	invalidJSON := `{not valid json}`

	_, err := ParseInvoice([]byte(invalidJSON))
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestInvoice_Validate(t *testing.T) {
	tests := []struct {
		name    string
		invoice Invoice
		wantErr bool
	}{
		{
			name: "valid invoice",
			invoice: Invoice{
				Meta:     Meta{InvoiceNumber: "INV-001", Date: "2025-01-10"},
				Sender:   Party{Name: "ACME Corp"},
				Customer: Party{Name: "John Doe"},
				Items:    []LineItem{{Description: "Service", Qty: 1, UnitPrice: 100}},
				Totals:   Totals{Net: 100, Tax: 0, Gross: 100},
			},
			wantErr: false,
		},
		{
			name: "missing invoice number",
			invoice: Invoice{
				Meta:     Meta{Date: "2025-01-10"},
				Sender:   Party{Name: "ACME Corp"},
				Customer: Party{Name: "John Doe"},
				Items:    []LineItem{{Description: "Service", Qty: 1, UnitPrice: 100}},
			},
			wantErr: true,
		},
		{
			name: "missing date",
			invoice: Invoice{
				Meta:     Meta{InvoiceNumber: "INV-001"},
				Sender:   Party{Name: "ACME Corp"},
				Customer: Party{Name: "John Doe"},
				Items:    []LineItem{{Description: "Service", Qty: 1, UnitPrice: 100}},
			},
			wantErr: true,
		},
		{
			name: "missing sender name",
			invoice: Invoice{
				Meta:     Meta{InvoiceNumber: "INV-001", Date: "2025-01-10"},
				Sender:   Party{},
				Customer: Party{Name: "John Doe"},
				Items:    []LineItem{{Description: "Service", Qty: 1, UnitPrice: 100}},
			},
			wantErr: true,
		},
		{
			name: "missing customer name",
			invoice: Invoice{
				Meta:     Meta{InvoiceNumber: "INV-001", Date: "2025-01-10"},
				Sender:   Party{Name: "ACME Corp"},
				Customer: Party{},
				Items:    []LineItem{{Description: "Service", Qty: 1, UnitPrice: 100}},
			},
			wantErr: true,
		},
		{
			name: "no items",
			invoice: Invoice{
				Meta:     Meta{InvoiceNumber: "INV-001", Date: "2025-01-10"},
				Sender:   Party{Name: "ACME Corp"},
				Customer: Party{Name: "John Doe"},
				Items:    []LineItem{},
			},
			wantErr: true,
		},
		{
			name: "item with empty description",
			invoice: Invoice{
				Meta:     Meta{InvoiceNumber: "INV-001", Date: "2025-01-10"},
				Sender:   Party{Name: "ACME Corp"},
				Customer: Party{Name: "John Doe"},
				Items:    []LineItem{{Description: "", Qty: 1, UnitPrice: 100}},
			},
			wantErr: true,
		},
		{
			name: "item with zero quantity",
			invoice: Invoice{
				Meta:     Meta{InvoiceNumber: "INV-001", Date: "2025-01-10"},
				Sender:   Party{Name: "ACME Corp"},
				Customer: Party{Name: "John Doe"},
				Items:    []LineItem{{Description: "Service", Qty: 0, UnitPrice: 100}},
			},
			wantErr: true,
		},
		{
			name: "item with negative price",
			invoice: Invoice{
				Meta:     Meta{InvoiceNumber: "INV-001", Date: "2025-01-10"},
				Sender:   Party{Name: "ACME Corp"},
				Customer: Party{Name: "John Doe"},
				Items:    []LineItem{{Description: "Service", Qty: 1, UnitPrice: -100}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.invoice.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLineItem_Amount(t *testing.T) {
	item := LineItem{Qty: 10, UnitPrice: 15.50}

	expected := 155.0
	if got := item.Amount(); got != expected {
		t.Errorf("Amount() = %v, want %v", got, expected)
	}
}

func TestLineItem_AmountWithVAT(t *testing.T) {
	item := LineItem{Qty: 10, UnitPrice: 100, VAT: 0.19}

	expected := 1190.0
	if got := item.AmountWithVAT(); got != expected {
		t.Errorf("AmountWithVAT() = %v, want %v", got, expected)
	}
}
