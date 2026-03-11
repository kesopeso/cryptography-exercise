package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kesopeso/cryptography-exercise/internal/bitset"
)

type PostgresStatusStore struct {
	db *pgx.Conn
}

func NewPostgresStatusStore(db *pgx.Conn) *PostgresStatusStore {
	return &PostgresStatusStore{db: db}
}

// Create inserts a new status row with UUID v7 and an empty encoded_status.
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
