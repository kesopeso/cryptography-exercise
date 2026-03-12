package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kesopeso/cryptography-exercise/internal/assert"
	"github.com/kesopeso/cryptography-exercise/internal/bitset"
	"github.com/kesopeso/cryptography-exercise/internal/cryptography"
)

// PostgresStatusStore implements StatusStore using a PostgreSQL connection.
type PostgresStatusStore struct {
	db          *pgx.Conn
	aesPassword string
}

// NewPostgresStatusStore creates a PostgresStatusStore with the given parameters.
func NewPostgresStatusStore(db *pgx.Conn, aesPassword string) *PostgresStatusStore {
	return &PostgresStatusStore{db: db, aesPassword: aesPassword}
}

// CreateStatus inserts a new status row with UUID v7 and encoded values from Bitset.
// Returns the generated UUID.
func (pss *PostgresStatusStore) CreateStatus(ctx context.Context, status *bitset.Bitset) (uuid.UUID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.UUID{}, err
	}

	encodedStatus, err := status.Encode()
	if err != nil {
		return uuid.UUID{}, err
	}

	encryptedStatus, err := cryptography.AESEncrypt(encodedStatus, pss.aesPassword)
	if err != nil {
		return uuid.UUID{}, err
	}

	_, err = pss.db.Exec(ctx, "INSERT INTO statuses (id, encoded_status, encrypted_status) VALUES ($1, $2, $3)", id, encodedStatus, encryptedStatus)
	if err != nil {
		return uuid.UUID{}, err
	}

	return id, nil
}

// GetStatusIds returns all status UUIDs from the database.
func (pss *PostgresStatusStore) GetStatusIds(ctx context.Context) ([]uuid.UUID, error) {
	rows, err := pss.db.Query(ctx, "SELECT id FROM statuses")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}

// GetEncodedStatus returns the encoded status string for the given statusId.
func (pss *PostgresStatusStore) GetEncodedStatus(ctx context.Context, statusId string) (string, error) {
	id, err := uuid.Parse(statusId)
	if err != nil {
		return "", fmt.Errorf("invalid status id: %w", err)
	}

	tx, err := pss.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var encodedStatus string
	var encryptedStatus []byte
	err = tx.QueryRow(ctx, "SELECT encoded_status, encrypted_status FROM statuses WHERE id = $1 FOR UPDATE", id).Scan(&encodedStatus, &encryptedStatus)
	if err != nil {
		return "", fmt.Errorf("failed to fetch statuses: %w", err)
	}

	if len(encryptedStatus) == 0 {
		encryptedStatus, err = cryptography.AESEncrypt(encodedStatus, pss.aesPassword)
		if err != nil {
			return "", fmt.Errorf("failed to encrypt encoded field: %w", err)
		}

		_, err = tx.Exec(ctx, "UPDATE statuses SET encrypted_status = $1 WHERE id = $2", encryptedStatus, id)
		if err != nil {
			return "", fmt.Errorf("failed to sync encrypted_status field: %w", err)
		}
	}

	derivedEncodedStatus, err := cryptography.AESDecrypt(encryptedStatus, pss.aesPassword)
	if err != nil {
		return "", fmt.Errorf("failed to derive encoded status: %w", err)
	}

	assert.True(encodedStatus == derivedEncodedStatus, fmt.Sprintf("encoded status %s, derived encoded status %s", encodedStatus, derivedEncodedStatus))

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return derivedEncodedStatus, nil
}

// CreateStatusValue adds new status value to the existing database status row.
// Returns the index of the newly added value.
func (pss *PostgresStatusStore) CreateStatusValue(ctx context.Context, statusId string, value bool) (int, error) {
	id, err := uuid.Parse(statusId)
	if err != nil {
		return -1, fmt.Errorf("invalid status id: %w", err)
	}

	tx, err := pss.db.Begin(ctx)
	if err != nil {
		return -1, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var encodedStatus string
	err = tx.QueryRow(ctx, "SELECT encoded_status FROM statuses WHERE id = $1 FOR UPDATE", id).Scan(&encodedStatus)
	if err != nil {
		return -1, fmt.Errorf("failed to fetch status: %w", err)
	}

	bs, err := bitset.Decode(encodedStatus)
	if err != nil {
		return -1, fmt.Errorf("failed to decode status: %w", err)
	}

	valueIndex := bs.Add(value)

	encodedBs, err := bs.Encode()
	if err != nil {
		return -1, fmt.Errorf("failed to encode status: %w", err)
	}

	encryptedBs, err := cryptography.AESEncrypt(encodedBs, pss.aesPassword)
	if err != nil {
		return -1, fmt.Errorf("failed to encrypt status: %w", err)
	}

	_, err = tx.Exec(ctx, "UPDATE statuses SET encoded_status = $1, encrypted_status = $2 WHERE id = $3", encodedBs, encryptedBs, id)
	if err != nil {
		return -1, fmt.Errorf("failed to update status row: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return -1, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return valueIndex, nil
}

// UpdateStatusValue updates the value at the given index in the status bitset.
// Uses a transaction with SELECT FOR UPDATE to prevent concurrent modifications.
func (pss *PostgresStatusStore) UpdateStatusValue(ctx context.Context, statusId string, index int, value bool) error {
	id, err := uuid.Parse(statusId)
	if err != nil {
		return fmt.Errorf("invalid status id: %w", err)
	}

	tx, err := pss.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var encodedStatus string
	err = tx.QueryRow(ctx, "SELECT encoded_status FROM statuses WHERE id = $1 FOR UPDATE", id).Scan(&encodedStatus)
	if err != nil {
		return fmt.Errorf("failed to fetch status: %w", err)
	}

	bs, err := bitset.Decode(encodedStatus)
	if err != nil {
		return fmt.Errorf("failed to decode status: %w", err)
	}

	if err := bs.Set(index, value); err != nil {
		return fmt.Errorf("failed to set status value: %w", err)
	}

	encodedBs, err := bs.Encode()
	if err != nil {
		return fmt.Errorf("failed to encode status: %w", err)
	}

	encryptedBs, err := cryptography.AESEncrypt(encodedBs, pss.aesPassword)
	if err != nil {
		return fmt.Errorf("failed to encrypt status: %w", err)
	}

	_, err = tx.Exec(ctx, "UPDATE statuses SET encoded_status = $1, encrypted_status = $2 WHERE id = $3", encodedBs, encryptedBs, id)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
