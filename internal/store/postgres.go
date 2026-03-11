package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kesopeso/cryptography-exercise/internal/bitset"
)

// PostgresStatusStore implements StatusStore using a PostgreSQL connection.
type PostgresStatusStore struct {
	db *pgx.Conn
}

// NewPostgresStatusStore creates a PostgresStatusStore with the given database connection.
func NewPostgresStatusStore(db *pgx.Conn) *PostgresStatusStore {
	return &PostgresStatusStore{db: db}
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

	_, err = pss.db.Exec(ctx, "INSERT INTO statuses (id, encoded_status) VALUES ($1, $2)", id, encodedStatus)
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

	_, err = tx.Exec(ctx, "UPDATE statuses SET encoded_status = $1 WHERE id = $2", encodedBs, id)

	if err := tx.Commit(ctx); err != nil {
		return -1, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return valueIndex, nil
}
