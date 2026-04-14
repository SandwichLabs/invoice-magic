package repository

import (
	"context"
	"errors"

	"github.com/sandwich-labs/invoice-generator-pro/internal/model"
)

// Sentinel errors returned by Repository implementations.
var (
	ErrNotFound    = errors.New("invoice not found")
	ErrDuplicateID = errors.New("invoice number already exists")
	ErrReadOnly    = errors.New("repository is read-only")
)

// Record is the canonical unit returned by every Repository method.
//
// RowNumber is a transitional field that preserves the existing {row}-based
// URL routing until issue #5 rekeys routes to {id}. It is 0 for backends
// that do not use row numbers.
type Record struct {
	ID        string         // equals Invoice.Meta.InvoiceNumber
	RowNumber int            // transitional; removed in #5
	Invoice   *model.Invoice
	RawJSON   []byte // pre-marshalled JSON ready for render.Renderer
}

// Repository is the storage abstraction for invoices.
// Implementations must be safe for concurrent use from a single goroutine
// (the HTTP server is single-process; full goroutine-safety is not required
// but is recommended).
type Repository interface {
	// List returns all invoices in the store, ordered as the backend stores them.
	List(ctx context.Context) ([]Record, error)

	// Get returns the invoice with the given ID (invoice_number).
	// Returns ErrNotFound if no match exists.
	Get(ctx context.Context, id string) (*Record, error)

	// Create stores a new invoice. Returns ErrDuplicateID if the invoice
	// number already exists. Auto-assigns an invoice number in the pattern
	// INV-YYYY-NNNN when inv.Meta.InvoiceNumber is blank.
	// Returns ErrReadOnly for read-only backends.
	Create(ctx context.Context, inv *model.Invoice) (*Record, error)

	// Update replaces the invoice identified by id with the supplied value.
	// Returns ErrNotFound if id does not exist.
	// Returns ErrReadOnly for read-only backends.
	Update(ctx context.Context, id string, inv *model.Invoice) (*Record, error)

	// Delete removes the invoice identified by id.
	// Returns ErrNotFound if id does not exist.
	// Returns ErrReadOnly for read-only backends.
	Delete(ctx context.Context, id string) error

	// Writable reports whether this repository supports mutations.
	// When false, Create/Update/Delete all return ErrReadOnly.
	Writable() bool
}
