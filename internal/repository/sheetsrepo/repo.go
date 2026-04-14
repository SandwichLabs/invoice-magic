// Package sheetsrepo implements repository.Repository backed by Google Sheets.
package sheetsrepo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sandwich-labs/invoice-generator-pro/internal/model"
	"github.com/sandwich-labs/invoice-generator-pro/internal/repository"
	"github.com/sandwich-labs/invoice-generator-pro/internal/sheets"
)

// Compile-time interface check.
var _ repository.Repository = (*Repo)(nil)

// Repo adapts sheets.Source to the repository.Repository interface.
// Create, Update, and Delete return repository.ErrReadOnly until issue #4
// implements the writer methods on sheets.Client.
type Repo struct {
	source   *sheets.Source
	writable bool
}

// New creates a Repo using the provided sheets client and source configuration.
// writable should be true only when the underlying HTTP client was granted the
// read/write Google Sheets scope (auth.AllScopes).
func New(client *sheets.Client, cfg sheets.SourceConfig, writable bool) *Repo {
	return &Repo{
		source:   sheets.NewSource(client, cfg),
		writable: writable,
	}
}

// Source returns the underlying sheets.Source for operations that are not part
// of the Repository interface, such as ProvisionHeaders.
func (r *Repo) Source() *sheets.Source {
	return r.source
}

// List returns all invoices from the configured sheet.
func (r *Repo) List(ctx context.Context) ([]repository.Record, error) {
	rows, err := r.source.FetchInvoices(ctx)
	if err != nil {
		return nil, err
	}

	records := make([]repository.Record, 0, len(rows))
	for _, row := range rows {
		rec, err := rowToRecord(row.RowNumber, row.Invoice, row.RawJSON)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", row.RowNumber, err)
		}
		records = append(records, rec)
	}
	return records, nil
}

// Get returns the invoice with the given ID (invoice_number).
// Because the Sheets API does not support filtering, this fetches all rows
// and searches linearly — acceptable for typical invoice sheet sizes.
func (r *Repo) Get(ctx context.Context, id string) (*repository.Record, error) {
	records, err := r.List(ctx)
	if err != nil {
		return nil, err
	}
	for i := range records {
		if records[i].ID == id {
			return &records[i], nil
		}
	}
	return nil, fmt.Errorf("%w: %s", repository.ErrNotFound, id)
}

// Create is not yet implemented for the sheets backend (issue #4).
func (r *Repo) Create(_ context.Context, _ *model.Invoice) (*repository.Record, error) {
	return nil, repository.ErrReadOnly
}

// Update is not yet implemented for the sheets backend (issue #4).
func (r *Repo) Update(_ context.Context, _ string, _ *model.Invoice) (*repository.Record, error) {
	return nil, repository.ErrReadOnly
}

// Delete is not yet implemented for the sheets backend (issue #4).
func (r *Repo) Delete(_ context.Context, _ string) error {
	return repository.ErrReadOnly
}

// Writable reports whether this repo was constructed with read/write scope.
func (r *Repo) Writable() bool {
	return r.writable
}

// rowToRecord converts a sheets InvoiceRow into a repository.Record.
func rowToRecord(rowNumber int, inv *model.Invoice, rawJSON []byte) (repository.Record, error) {
	if rawJSON == nil {
		var err error
		rawJSON, err = json.Marshal(inv)
		if err != nil {
			return repository.Record{}, fmt.Errorf("marshal invoice: %w", err)
		}
	}
	return repository.Record{
		ID:        inv.Meta.InvoiceNumber,
		RowNumber: rowNumber,
		Invoice:   inv,
		RawJSON:   rawJSON,
	}, nil
}
