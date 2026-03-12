package service

import (
	"context"

	"github.com/google/uuid"
)

// StatusService is the interface that our HTTP handlers will use
type StatusService interface {
	CreateStatus(ctx context.Context, status []bool) (uuid.UUID, error)
	GetStatusIds(ctx context.Context) ([]uuid.UUID, error)
	GetEncodedStatus(ctx context.Context, statusId string) (string, error)
	CreateStatusValue(ctx context.Context, statusId string, value bool) (int, error)
	UpdateStatusValue(ctx context.Context, statusId string, index int, value bool) error
}
