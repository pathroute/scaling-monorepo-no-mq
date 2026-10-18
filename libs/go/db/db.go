// Package db is the storage layer services share: a typed table with transactions.
// The in-memory backend stands in for Postgres; the interface is what services code against.
package db

import (
	"cmp"
	"errors"
	"maps"
	"slices"
	"sync"

	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("db: not found")
	ErrExists   = errors.New("db: already exists")
)

// Table is a keyed collection of rows of type T.
type Table[T any] struct {
	mu   sync.RWMutex
	rows map[string]T
}

// NewTable returns an empty table.
func NewTable[T any]() *Table[T] { return &Table[T]{rows: map[string]T{}} }

// NewID returns a random row id.
func NewID() string { return uuid.NewString() }

// Get returns the row with id.
func (t *Table[T]) Get(id string) (T, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	v, ok := t.rows[id]
	if !ok {
		var zero T
		return zero, ErrNotFound
	}
	return v, nil
}

// Insert adds a row; the id must be new.
func (t *Table[T]) Insert(id string, v T) error {
	return t.Tx(func(tx *Tx[T]) error { return tx.Insert(id, v) })
}

// Delete removes a row.
func (t *Table[T]) Delete(id string) error {
	return t.Tx(func(tx *Tx[T]) error { return tx.Delete(id) })
}

// List returns the rows matching keep (all rows if keep is nil), ordered by id.
func (t *Table[T]) List(keep func(T) bool) []T {
	t.mu.RLock()
	defer t.mu.RUnlock()
	ids := slices.SortedFunc(maps.Keys(t.rows), cmp.Compare[string])
	out := make([]T, 0, len(ids))
	for _, id := range ids {
		if v := t.rows[id]; keep == nil || keep(v) {
			out = append(out, v)
		}
	}
	return out
}

// Tx runs fn atomically: if it returns an error, none of its writes apply.
func (t *Table[T]) Tx(fn func(*Tx[T]) error) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	tx := &Tx[T]{base: t.rows, writes: map[string]*T{}}
	if err := fn(tx); err != nil {
		return err
	}
	for id, v := range tx.writes {
		if v == nil {
			delete(t.rows, id)
		} else {
			t.rows[id] = *v
		}
	}
	return nil
}

// Tx is a write set over a table; a nil entry marks a delete.
type Tx[T any] struct {
	base   map[string]T
	writes map[string]*T
}

// Get reads through the transaction's own writes.
func (tx *Tx[T]) Get(id string) (T, error) {
	if v, ok := tx.writes[id]; ok {
		if v == nil {
			var zero T
			return zero, ErrNotFound
		}
		return *v, nil
	}
	v, ok := tx.base[id]
	if !ok {
		var zero T
		return zero, ErrNotFound
	}
	return v, nil
}

// Insert adds a row; the id must be new.
func (tx *Tx[T]) Insert(id string, v T) error {
	if _, err := tx.Get(id); err == nil {
		return ErrExists
	}
	tx.writes[id] = &v
	return nil
}

// Update replaces an existing row.
func (tx *Tx[T]) Update(id string, v T) error {
	if _, err := tx.Get(id); err != nil {
		return err
	}
	tx.writes[id] = &v
	return nil
}

// Delete removes an existing row.
func (tx *Tx[T]) Delete(id string) error {
	if _, err := tx.Get(id); err != nil {
		return err
	}
	tx.writes[id] = nil
	return nil
}
