package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/kesopeso/cryptography-exercise/internal/bitset"
)

// StatusStore is the interface that our HTTP handlers will use
type StatusStore interface {
	CreateStatus(ctx context.Context, status *bitset.Bitset) (uuid.UUID, error)
	GetStatusIds(ctx context.Context) ([]uuid.UUID, error)
}
