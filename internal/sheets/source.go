package sheets

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/sandwich-labs/invoice-generator-pro/internal/model"
)

// FieldMapping defines how sheet columns map to invoice fields.
type FieldMapping struct {
	// Column name or letter (e.g., "A", "Invoice Number", etc.)
	Column string `yaml:"column" json:"column"`
}

// SourceConfig defines the configuration for reading invoices from Google Sheets.
type SourceConfig struct {
	SpreadsheetID string `yaml:"spreadsheet_id" json:"spreadsheet_id"`
	SheetName     string `yaml:"sheet_name" json:"sheet_name"`
	HeaderRow     int    `yaml:"header_row" json:"header_row"` // 1-indexed, default 1
	DataStartRow  int    `yaml:"data_start_row" json:"data_start_row"` // 1-indexed, default 2

	// Field mappings - map invoice fields to column names/letters
	Fields FieldMappings `yaml:"fields" json:"fields"`
}

// FieldMappings contains all configurable field mappings.
type FieldMappings struct {
	// Meta fields
	InvoiceNumber string `yaml:"invoice_number" json:"invoice_number"`
	Date          string `yaml:"date" json:"date"`
	DueDate       string `yaml:"due_date" json:"due_date"`
	Currency      string `yaml:"currency" json:"currency"`

	// Sender fields
	SenderName    string `yaml:"sender_name" json:"sender_name"`
	SenderCompany string `yaml:"sender_company" json:"sender_company"`
	SenderAddress string `yaml:"sender_address" json:"sender_address"`
	SenderTaxID   string `yaml:"sender_tax_id" json:"sender_tax_id"`
	SenderEmail   string `yaml:"sender_email" json:"sender_email"`
	SenderPhone   string `yaml:"sender_phone" json:"sender_phone"`

	// Customer fields
	CustomerName    string `yaml:"customer_name" json:"customer_name"`
	CustomerCompany string `yaml:"customer_company" json:"customer_company"`
	CustomerAddress string `yaml:"customer_address" json:"customer_address"`
	CustomerTaxID   string `yaml:"customer_tax_id" json:"customer_tax_id"`
	CustomerEmail   string `yaml:"customer_email" json:"customer_email"`
	CustomerPhone   string `yaml:"customer_phone" json:"customer_phone"`

	// Line item fields (for single-item invoices from sheets)
	ItemDescription string `yaml:"item_description" json:"item_description"`
	ItemQty         string `yaml:"item_qty" json:"item_qty"`
	ItemUnitPrice   string `yaml:"item_unit_price" json:"item_unit_price"`
	ItemVAT         string `yaml:"item_vat" json:"item_vat"`

	// Totals (optional - can be calculated if not provided)
	TotalNet   string `yaml:"total_net" json:"total_net"`
	TotalTax   string `yaml:"total_tax" json:"total_tax"`
	TotalGross string `yaml:"total_gross" json:"total_gross"`

	// Notes
	Notes string `yaml:"notes" json:"notes"`
}

// Source reads invoice data from Google Sheets.
type Source struct {
	client *Client
	config SourceConfig
}

// NewSource creates a new sheets data source.
func NewSource(client *Client, config SourceConfig) *Source {
	if config.HeaderRow == 0 {
		config.HeaderRow = 1
	}
	if config.DataStartRow == 0 {
		config.DataStartRow = 2
	}
	return &Source{
		client: client,
		config: config,
	}
}

// InvoiceRow represents a single invoice extracted from a sheet row.
type InvoiceRow struct {
	RowNumber int
	Invoice   *model.Invoice
	RawJSON   []byte
}

// FetchInvoices retrieves all invoices from the configured sheet.
func (s *Source) FetchInvoices(ctx context.Context) ([]InvoiceRow, error) {
	data, err := s.client.GetSheetData(ctx, s.config.SpreadsheetID, s.config.SheetName)
	if err != nil {
		return nil, fmt.Errorf("fetch sheet data: %w", err)
	}

	if len(data) < s.config.HeaderRow {
		return nil, fmt.Errorf("sheet has fewer rows than header_row setting")
	}

	// Build column name to index mapping from header row
	headerRow := data[s.config.HeaderRow-1]
	colIndex := make(map[string]int)
	for i, cell := range headerRow {
		if name, ok := cell.(string); ok {
			colIndex[strings.TrimSpace(strings.ToLower(name))] = i
			// Also index by column letter
			colIndex[columnLetter(i)] = i
		}
	}

	var invoices []InvoiceRow
	for rowIdx := s.config.DataStartRow - 1; rowIdx < len(data); rowIdx++ {
		row := data[rowIdx]
		if isEmptyRow(row) {
			continue
		}

		invoice, err := s.rowToInvoice(row, colIndex)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", rowIdx+1, err)
		}

		rawJSON, err := json.Marshal(invoice)
		if err != nil {
			return nil, fmt.Errorf("row %d: marshal invoice: %w", rowIdx+1, err)
		}

		invoices = append(invoices, InvoiceRow{
			RowNumber: rowIdx + 1,
			Invoice:   invoice,
			RawJSON:   rawJSON,
		})
	}

	return invoices, nil
}

// FetchInvoice retrieves a single invoice by row number (1-indexed).
func (s *Source) FetchInvoice(ctx context.Context, rowNumber int) (*InvoiceRow, error) {
	invoices, err := s.FetchInvoices(ctx)
	if err != nil {
		return nil, err
	}

	for _, inv := range invoices {
		if inv.RowNumber == rowNumber {
			return &inv, nil
		}
	}

	return nil, fmt.Errorf("no invoice found at row %d", rowNumber)
}

func (s *Source) rowToInvoice(row []interface{}, colIndex map[string]int) (*model.Invoice, error) {
	f := s.config.Fields

	invoice := &model.Invoice{
		Meta: model.Meta{
			InvoiceNumber: s.getString(row, colIndex, f.InvoiceNumber),
			Date:          s.getString(row, colIndex, f.Date),
			DueDate:       s.getString(row, colIndex, f.DueDate),
			Currency:      s.getString(row, colIndex, f.Currency),
		},
		Sender: model.Party{
			Name:    s.getString(row, colIndex, f.SenderName),
			Company: s.getString(row, colIndex, f.SenderCompany),
			Address: s.getString(row, colIndex, f.SenderAddress),
			TaxID:   s.getString(row, colIndex, f.SenderTaxID),
			Email:   s.getString(row, colIndex, f.SenderEmail),
			Phone:   s.getString(row, colIndex, f.SenderPhone),
		},
		Customer: model.Party{
			Name:    s.getString(row, colIndex, f.CustomerName),
			Company: s.getString(row, colIndex, f.CustomerCompany),
			Address: s.getString(row, colIndex, f.CustomerAddress),
			TaxID:   s.getString(row, colIndex, f.CustomerTaxID),
			Email:   s.getString(row, colIndex, f.CustomerEmail),
			Phone:   s.getString(row, colIndex, f.CustomerPhone),
		},
		Notes: s.getString(row, colIndex, f.Notes),
	}

	// Handle line item
	if f.ItemDescription != "" {
		item := model.LineItem{
			Description: s.getString(row, colIndex, f.ItemDescription),
			Qty:         s.getFloat(row, colIndex, f.ItemQty, 1),
			UnitPrice:   s.getFloat(row, colIndex, f.ItemUnitPrice, 0),
			VAT:         s.getFloat(row, colIndex, f.ItemVAT, 0),
		}
		invoice.Items = []model.LineItem{item}
	}

	// Handle totals
	if f.TotalNet != "" || f.TotalTax != "" || f.TotalGross != "" {
		invoice.Totals = model.Totals{
			Net:   s.getFloat(row, colIndex, f.TotalNet, 0),
			Tax:   s.getFloat(row, colIndex, f.TotalTax, 0),
			Gross: s.getFloat(row, colIndex, f.TotalGross, 0),
		}
	} else if len(invoice.Items) > 0 {
		// Calculate totals from items
		var net, tax float64
		for _, item := range invoice.Items {
			itemNet := item.Amount()
			net += itemNet
			tax += itemNet * item.VAT
		}
		invoice.Totals = model.Totals{
			Net:   net,
			Tax:   tax,
			Gross: net + tax,
		}
	}

	return invoice, nil
}

func (s *Source) getString(row []interface{}, colIndex map[string]int, colName string) string {
	if colName == "" {
		return ""
	}

	idx, ok := colIndex[strings.ToLower(colName)]
	if !ok {
		// Try as column letter
		idx, ok = colIndex[strings.ToUpper(colName)]
		if !ok {
			return ""
		}
	}

	if idx >= len(row) {
		return ""
	}

	switch v := row[idx].(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (s *Source) getFloat(row []interface{}, colIndex map[string]int, colName string, defaultVal float64) float64 {
	if colName == "" {
		return defaultVal
	}

	idx, ok := colIndex[strings.ToLower(colName)]
	if !ok {
		idx, ok = colIndex[strings.ToUpper(colName)]
		if !ok {
			return defaultVal
		}
	}

	if idx >= len(row) {
		return defaultVal
	}

	switch v := row[idx].(type) {
	case float64:
		return v
	case string:
		v = strings.TrimSpace(v)
		// Remove currency symbols and commas
		v = strings.ReplaceAll(v, ",", "")
		v = strings.ReplaceAll(v, "$", "")
		v = strings.ReplaceAll(v, "€", "")
		v = strings.ReplaceAll(v, "£", "")
		v = strings.TrimSpace(v)
		if v == "" {
			return defaultVal
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return defaultVal
		}
		return f
	default:
		return defaultVal
	}
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

func isEmptyRow(row []interface{}) bool {
	for _, cell := range row {
		switch v := cell.(type) {
		case string:
			if strings.TrimSpace(v) != "" {
				return false
			}
		case nil:
			continue
		default:
			return false
		}
	}
	return true
}

// ProvisionHeaders writes the configured field mapping headers to the sheet's header row.
// This sets up the sheet structure based on the field mappings in config.
// Requires spreadsheets (read/write) scope.
func (s *Source) ProvisionHeaders(ctx context.Context) error {
	headers := s.buildHeaderRow()
	if len(headers) == 0 {
		return fmt.Errorf("no field mappings configured")
	}

	// Verify the sheet exists
	sheet, err := s.client.FindSheetByName(ctx, s.config.SpreadsheetID, s.config.SheetName)
	if err != nil {
		return fmt.Errorf("check sheet exists: %w", err)
	}
	if sheet == nil {
		return fmt.Errorf("sheet %q not found in spreadsheet. Create the sheet first or check the sheet_name in config.yaml", s.config.SheetName)
	}

	// Build range for header row (e.g., "'Invoices'!A1:T1")
	// Sheet names must be quoted in A1 notation
	endCol := columnLetter(len(headers) - 1)
	rangeA1 := fmt.Sprintf("'%s'!A%d:%s%d",
		s.config.SheetName,
		s.config.HeaderRow,
		endCol,
		s.config.HeaderRow,
	)

	values := [][]interface{}{headers}
	return s.client.UpdateRange(ctx, s.config.SpreadsheetID, rangeA1, values)
}

// buildHeaderRow creates the header row values from field mappings.
func (s *Source) buildHeaderRow() []interface{} {
	f := s.config.Fields

	// Collect all non-empty field mappings in order
	fieldOrder := []string{
		f.InvoiceNumber,
		f.Date,
		f.DueDate,
		f.Currency,
		f.SenderName,
		f.SenderCompany,
		f.SenderAddress,
		f.SenderTaxID,
		f.SenderEmail,
		f.SenderPhone,
		f.CustomerName,
		f.CustomerCompany,
		f.CustomerAddress,
		f.CustomerTaxID,
		f.CustomerEmail,
		f.CustomerPhone,
		f.ItemDescription,
		f.ItemQty,
		f.ItemUnitPrice,
		f.ItemVAT,
		f.TotalNet,
		f.TotalTax,
		f.TotalGross,
		f.Notes,
	}

	var headers []interface{}
	for _, h := range fieldOrder {
		if h != "" {
			headers = append(headers, h)
		}
	}

	return headers
}

// GetHeaderMapping returns the configured field names for display.
func (s *Source) GetHeaderMapping() []string {
	headers := s.buildHeaderRow()
	result := make([]string, len(headers))
	for i, h := range headers {
		result[i] = h.(string)
	}
	return result
}
