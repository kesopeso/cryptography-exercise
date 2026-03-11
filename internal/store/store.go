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
	GetEncodedStatus(ctx context.Context, statusId string) (string, error)
	CreateStatusValue(ctx context.Context, statusId string, value bool) (int, error)
	UpdateStatusValue(ctx context.Context, statusId string, index int, value bool) error
}
