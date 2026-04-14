// Package memrepo provides an in-memory Repository implementation for local
// development and testing. It is not intended for production use.
package memrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sandwich-labs/invoice-generator-pro/internal/model"
	"github.com/sandwich-labs/invoice-generator-pro/internal/repository"
)

// Compile-time interface check.
var _ repository.Repository = (*Repo)(nil)

// Repo is a thread-safe in-memory repository.
type Repo struct {
	mu      sync.RWMutex
	records []repository.Record // insertion-ordered
	nextRow int
}

// New creates an empty in-memory repository.
func New() *Repo {
	return &Repo{nextRow: 1}
}

// Seed adds an invoice directly, bypassing validation. Useful for loading
// test fixtures at startup.
func (r *Repo) Seed(inv *model.Invoice) error {
	raw, err := json.Marshal(inv)
	if err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = append(r.records, repository.Record{
		ID:        inv.Meta.InvoiceNumber,
		RowNumber: r.nextRow,
		Invoice:   inv,
		RawJSON:   raw,
	})
	r.nextRow++
	return nil
}

func (r *Repo) List(_ context.Context) ([]repository.Record, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]repository.Record, len(r.records))
	copy(out, r.records)
	return out, nil
}

func (r *Repo) Get(_ context.Context, id string) (*repository.Record, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for i := range r.records {
		if r.records[i].ID == id {
			rec := r.records[i]
			return &rec, nil
		}
	}
	return nil, fmt.Errorf("%w: %s", repository.ErrNotFound, id)
}

func (r *Repo) Create(_ context.Context, inv *model.Invoice) (*repository.Record, error) {
	if inv.Meta.InvoiceNumber == "" {
		inv.Meta.InvoiceNumber = fmt.Sprintf("INV-%s-%04d",
			time.Now().Format("2006"), r.nextRow)
	}
	raw, err := json.Marshal(inv)
	if err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, rec := range r.records {
		if rec.ID == inv.Meta.InvoiceNumber {
			return nil, fmt.Errorf("%w: %s", repository.ErrDuplicateID, inv.Meta.InvoiceNumber)
		}
	}
	rec := repository.Record{
		ID:        inv.Meta.InvoiceNumber,
		RowNumber: r.nextRow,
		Invoice:   inv,
		RawJSON:   raw,
	}
	r.records = append(r.records, rec)
	r.nextRow++
	return &rec, nil
}

func (r *Repo) Update(_ context.Context, id string, inv *model.Invoice) (*repository.Record, error) {
	raw, err := json.Marshal(inv)
	if err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.records {
		if r.records[i].ID == id {
			inv.Meta.InvoiceNumber = id
			r.records[i].Invoice = inv
			r.records[i].RawJSON = raw
			rec := r.records[i]
			return &rec, nil
		}
	}
	return nil, fmt.Errorf("%w: %s", repository.ErrNotFound, id)
}

func (r *Repo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.records {
		if r.records[i].ID == id {
			r.records = append(r.records[:i], r.records[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("%w: %s", repository.ErrNotFound, id)
}

func (r *Repo) Writable() bool { return true }
