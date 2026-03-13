package store

import (
	"context"

	"github.com/google/uuid"
)

// TransactorFn is a function that executes within a database transaction.
// The provided context carries the active transaction and must be forwarded
// to any store calls made inside fn.
type TransactorFn func(context.Context) error

// Transactor wraps database operations in an atomic transaction.
type Transactor interface {
	WithTransaction(ctx context.Context, fn TransactorFn) error
}

// Status represents a single status record as stored in the database.
type Status struct {
	Id              uuid.UUID
	EncodedStatus   string
	EncryptedStatus []byte
}

// StatusStore is the primary data access interface for status records.
// All mutating operations participate in transactions when the context
// provided to them carries an active transaction (see Transactor).
type StatusStore interface {
	Transactor
	InsertStatus(ctx context.Context, encodedStatus string, encryptedStatus []byte) (uuid.UUID, error)
	GetStatusIds(ctx context.Context) ([]uuid.UUID, error)
	GetStatus(ctx context.Context, id uuid.UUID) (Status, error)
	UpdateEncryptedStatus(ctx context.Context, id uuid.UUID, encryptedStatus []byte) error
	GetEncodedStatus(ctx context.Context, id uuid.UUID) (string, error)
	UpdateStatus(ctx context.Context, status Status) error
}
