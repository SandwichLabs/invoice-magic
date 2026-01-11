package model

import (
	"encoding/json"
	"fmt"
)

// Invoice represents the complete invoice data structure
type Invoice struct {
	Meta     Meta       `json:"meta"`
	Sender   Party      `json:"sender"`
	Customer Party      `json:"customer"`
	Items    []LineItem `json:"items"`
	Totals   Totals     `json:"totals"`
	Notes    string     `json:"notes,omitempty"`
}

// Meta contains invoice metadata
type Meta struct {
	InvoiceNumber string `json:"invoice_number"`
	Date          string `json:"date"`
	DueDate       string `json:"due_date,omitempty"`
	Currency      string `json:"currency,omitempty"`
}

// Party represents a sender or customer
type Party struct {
	Name    string `json:"name"`
	Company string `json:"company,omitempty"`
	Address string `json:"address,omitempty"`
	TaxID   string `json:"tax_id,omitempty"`
	Email   string `json:"email,omitempty"`
	Phone   string `json:"phone,omitempty"`
}

// LineItem represents a single invoice line item
type LineItem struct {
	Description string  `json:"description"`
	Qty         float64 `json:"qty"`
	UnitPrice   float64 `json:"unit_price"`
	VAT         float64 `json:"vat"`
}

// Amount calculates the total amount for this line item (before VAT)
func (li LineItem) Amount() float64 {
	return li.Qty * li.UnitPrice
}

// AmountWithVAT calculates the total amount including VAT
func (li LineItem) AmountWithVAT() float64 {
	return li.Amount() * (1 + li.VAT)
}

// Totals contains the invoice totals
type Totals struct {
	Net   float64 `json:"net"`
	Tax   float64 `json:"tax"`
	Gross float64 `json:"gross"`
}

// ParseInvoice parses JSON data into an Invoice struct
func ParseInvoice(data []byte) (*Invoice, error) {
	var invoice Invoice
	if err := json.Unmarshal(data, &invoice); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return &invoice, nil
}

// Validate checks that the invoice contains all required fields
func (inv *Invoice) Validate() error {
	if inv.Meta.InvoiceNumber == "" {
		return fmt.Errorf("missing required field: meta.invoice_number")
	}
	if inv.Meta.Date == "" {
		return fmt.Errorf("missing required field: meta.date")
	}
	if inv.Sender.Name == "" {
		return fmt.Errorf("missing required field: sender.name")
	}
	if inv.Customer.Name == "" {
		return fmt.Errorf("missing required field: customer.name")
	}
	if len(inv.Items) == 0 {
		return fmt.Errorf("invoice must have at least one line item")
	}

	for i, item := range inv.Items {
		if item.Description == "" {
			return fmt.Errorf("line item %d: missing description", i+1)
		}
		if item.Qty <= 0 {
			return fmt.Errorf("line item %d: quantity must be positive", i+1)
		}
		if item.UnitPrice < 0 {
			return fmt.Errorf("line item %d: unit price cannot be negative", i+1)
		}
		if item.VAT < 0 {
			return fmt.Errorf("line item %d: VAT rate cannot be negative", i+1)
		}
	}

	return nil
}
