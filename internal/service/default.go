package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/kesopeso/cryptography-exercise/internal/assert"
	"github.com/kesopeso/cryptography-exercise/internal/bitset"
	"github.com/kesopeso/cryptography-exercise/internal/store"
	"github.com/kesopeso/cryptography-exercise/pkg/cryptography"
)

// DefaultStatusService implements StatusService.
type DefaultStatusService struct {
	statusStore store.StatusStore
	aesPassword string
}

// NewDefaultStatusService creates a DefaultStatusService with the given parameters.
func NewDefaultStatusService(statusStore store.StatusStore, aesPassword string) *DefaultStatusService {
	return &DefaultStatusService{statusStore: statusStore, aesPassword: aesPassword}
}

// CreateStatus inserts a new status row with encoded and encrypted values from status parameter.
// Returns the generated UUID.
func (dss *DefaultStatusService) CreateStatus(ctx context.Context, status []bool) (uuid.UUID, error) {
	bs := bitset.NewBitset()
	for _, v := range status {
		bs.Add(v)
	}

	encodedStatus, err := bs.Encode()
	if err != nil {
		return uuid.UUID{}, err
	}

	encryptedStatus, err := cryptography.AESEncrypt(encodedStatus, dss.aesPassword)
	if err != nil {
		return uuid.UUID{}, err
	}

	id, err := dss.statusStore.InsertStatus(ctx, encodedStatus, encryptedStatus)
	if err != nil {
		return uuid.UUID{}, err
	}

	return id, nil
}

// GetStatusIds returns all status UUIDs from the database.
func (dss *DefaultStatusService) GetStatusIds(ctx context.Context) ([]uuid.UUID, error) {
	return dss.statusStore.GetStatusIds(ctx)
}

// GetEncodedStatus returns the encoded status string for the given statusId.
func (dss *DefaultStatusService) GetEncodedStatus(ctx context.Context, statusId string) (string, error) {
	id, err := uuid.Parse(statusId)
	if err != nil {
		return "", fmt.Errorf("invalid status id: %w", err)
	}

	var derivedEncodedStatus string
	err = dss.statusStore.WithTransaction(ctx, func(ctx context.Context) error {
		status, err := dss.statusStore.GetStatus(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to fetch statuses: %w", err)
		}

		// this is part of the code is unnecessary
		// it is just placed here for the fun of it,
		// to demonstrate that encryption / decryption works
		if len(status.EncryptedStatus) == 0 {
			status.EncryptedStatus, err = cryptography.AESEncrypt(status.EncodedStatus, dss.aesPassword)
			if err != nil {
				return fmt.Errorf("failed to encrypt encoded status: %w", err)
			}

			err := dss.statusStore.UpdateEncryptedStatus(ctx, id, status.EncryptedStatus)
			if err != nil {
				return fmt.Errorf("failed to sync encrypted_status field: %w", err)
			}
		}

		derivedEncodedStatus, err = cryptography.AESDecrypt(status.EncryptedStatus, dss.aesPassword)
		if err != nil {
			return fmt.Errorf("failed to derive encoded status from encrypted status: %w", err)
		}

		assert.True(
			status.EncodedStatus == derivedEncodedStatus,
			fmt.Sprintf(
				"encoded status %s, derived encoded status %s",
				status.EncodedStatus,
				derivedEncodedStatus,
			),
		)

		return nil
	})

	return derivedEncodedStatus, err
}

// CreateStatusValue adds new status value to the existing database status row.
// Returns the index of the newly added value.
func (dss *DefaultStatusService) CreateStatusValue(ctx context.Context, statusId string, value bool) (int, error) {
	id, err := uuid.Parse(statusId)
	if err != nil {
		return -1, fmt.Errorf("invalid status id: %w", err)
	}

	valueIndex := -1
	err = dss.statusStore.WithTransaction(ctx, func(ctx context.Context) error {
		encodedStatus, err := dss.statusStore.GetEncodedStatus(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to fetch status: %w", err)
		}

		bs, err := bitset.Decode(encodedStatus)
		if err != nil {
			return fmt.Errorf("failed to decode status: %w", err)
		}

		valueIndex = bs.Add(value)

		encodedBs, err := bs.Encode()
		if err != nil {
			return fmt.Errorf("failed to encode status: %w", err)
		}

		encryptedBs, err := cryptography.AESEncrypt(encodedBs, dss.aesPassword)
		if err != nil {
			return fmt.Errorf("failed to encrypt status: %w", err)
		}

		status := store.Status{Id: id, EncodedStatus: encodedBs, EncryptedStatus: encryptedBs}
		err = dss.statusStore.UpdateStatus(ctx, status)
		if err != nil {
			return fmt.Errorf("failed to update status row: %w", err)
		}

		return nil
	})

	return valueIndex, err
}

// UpdateStatusValue updates the value at the given index in the status bitset.
// Uses a transaction with SELECT FOR UPDATE to prevent concurrent modifications.
func (dss *DefaultStatusService) UpdateStatusValue(ctx context.Context, statusId string, index int, value bool) error {
	id, err := uuid.Parse(statusId)
	if err != nil {
		return fmt.Errorf("invalid status id: %w", err)
	}

	return dss.statusStore.WithTransaction(ctx, func(ctx context.Context) error {
		encodedStatus, err := dss.statusStore.GetEncodedStatus(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to fetch status: %w", err)
		}

		bs, err := bitset.Decode(encodedStatus)
		if err != nil {
			return fmt.Errorf("failed to decode status: %w", err)
		}

		if err = bs.Set(index, value); err != nil {
			return fmt.Errorf("failed to set status value: %w", err)
		}

		encodedBs, err := bs.Encode()
		if err != nil {
			return fmt.Errorf("failed to encode status: %w", err)
		}

		encryptedBs, err := cryptography.AESEncrypt(encodedBs, dss.aesPassword)
		if err != nil {
			return fmt.Errorf("failed to encrypt status: %w", err)
		}

		status := store.Status{Id: id, EncodedStatus: encodedBs, EncryptedStatus: encryptedBs}
		err = dss.statusStore.UpdateStatus(ctx, status)
		if err != nil {
			return fmt.Errorf("failed to update status row: %w", err)
		}

		return nil
	})
}
