package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// PostgresStatusStore is a PostgreSQL-backed implementation of StatusStore.
type PostgresStatusStore struct {
	db *pgx.Conn
}

// NewPostgresStatusStore creates a PostgresStatusStore using the given connection.
func NewPostgresStatusStore(db *pgx.Conn) *PostgresStatusStore {
	return &PostgresStatusStore{db: db}
}

// WithTransaction begins a transaction, injects it into the context, calls fn,
// and commits on success or rolls back on error or panic.
func (ps *PostgresStatusStore) WithTransaction(ctx context.Context, fn TransactorFn) error {
	tx, err := ps.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	ctx = createContextWithTx(ctx, tx)
	if err := fn(ctx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// InsertStatus persists a new status record and returns its generated UUIDv7.
func (pss *PostgresStatusStore) InsertStatus(ctx context.Context, encodedStatus string, encryptedStatus []byte) (uuid.UUID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.UUID{}, err
	}

	sql := `
		INSERT INTO statuses (
			id,
			encoded_status,
			encrypted_status
		) VALUES (
			$1,
			$2,
			$3
		)
	`

	execFn, _ := pss.getExecFn(ctx)

	_, err = execFn(ctx, sql, id, encodedStatus, encryptedStatus)
	if err != nil {
		return uuid.UUID{}, err
	}

	return id, nil
}

// GetStatusIds returns the IDs of all status records.
func (ps *PostgresStatusStore) GetStatusIds(ctx context.Context) ([]uuid.UUID, error) {
	queryFn, _ := ps.getQueryFn(ctx)
	rows, err := queryFn(ctx, "SELECT id FROM statuses")
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
// GetStatus retrieves a full status record by ID. The row is locked with
// SELECT FOR UPDATE when called inside a transaction.
func (pss *PostgresStatusStore) GetStatus(ctx context.Context, id uuid.UUID) (Status, error) {
	queryRowFn, isTxSet := pss.getQueryRowFn(ctx)

	sql := `
		SELECT
			encoded_status,
			encrypted_status
		FROM statuses
		WHERE id = $1
	`
	if isTxSet {
		sql += " FOR UPDATE"
	}

	var encodedStatus string
	var encryptedStatus []byte
	err := queryRowFn(ctx, sql, id).Scan(&encodedStatus, &encryptedStatus)
	if err != nil {
		return Status{}, err
	}

	status := Status{
		Id:              id,
		EncodedStatus:   encodedStatus,
		EncryptedStatus: encryptedStatus,
	}

	return status, nil
}

// UpdateEncryptedStatus replaces the encrypted_status field for the given ID.
func (pss *PostgresStatusStore) UpdateEncryptedStatus(ctx context.Context, id uuid.UUID, encryptedStatus []byte) error {
	execFn, _ := pss.getExecFn(ctx)

	sql := `
		UPDATE statuses
		SET encrypted_status = $1
		WHERE id = $2
	`

	_, err := execFn(ctx, sql, encryptedStatus, id)
	return err
}
// GetEncodedStatus retrieves the encoded_status field for the given ID. The row
// is locked with SELECT FOR UPDATE when called inside a transaction.
func (pss *PostgresStatusStore) GetEncodedStatus(ctx context.Context, id uuid.UUID) (string, error) {
	queryRowFn, isTxSet := pss.getQueryRowFn(ctx)

	sql := `
		SELECT encoded_status
		FROM statuses
		WHERE id = $1
	`

	if isTxSet {
		sql += " FOR UPDATE"
	}

	var encodedStatus string
	err := queryRowFn(ctx, sql, id).Scan(&encodedStatus)

	return encodedStatus, err
}

// UpdateStatus replaces both encoded_status and encrypted_status for the record
// identified by status.Id.
func (pss *PostgresStatusStore) UpdateStatus(ctx context.Context, status Status) error {
	execFn, _ := pss.getExecFn(ctx)

	sql := `
		UPDATE statuses
		SET
			encoded_status = $1,
			encrypted_status = $2
		WHERE id = $3
	`

	_, err := execFn(ctx, sql, status.EncodedStatus, status.EncryptedStatus, status.Id)
	return err
}

type execFn func(context.Context, string, ...any) (pgconn.CommandTag, error)

// getExecFn returns the Exec function for the active transaction in ctx, or
// falls back to the connection-level Exec.
func (pss *PostgresStatusStore) getExecFn(ctx context.Context) (execFn execFn, isTxSet bool) {
	tx := getTxFromContext(ctx)
	if tx == nil {
		execFn = pss.db.Exec
		isTxSet = false
		return
	}

	execFn = tx.Exec
	isTxSet = true
	return
}

type queryFn func(context.Context, string, ...any) (pgx.Rows, error)

// getQueryFn returns the Query function for the active transaction in ctx, or
// falls back to the connection-level Query.
func (pss *PostgresStatusStore) getQueryFn(ctx context.Context) (queryFn queryFn, isTxSet bool) {
	tx := getTxFromContext(ctx)
	if tx == nil {
		queryFn = pss.db.Query
		isTxSet = false
		return
	}

	queryFn = tx.Query
	isTxSet = true
	return
}

type queryRowFn func(context.Context, string, ...any) pgx.Row

// getQueryRowFn returns the QueryRow function for the active transaction in ctx,
// or falls back to the connection-level QueryRow.
func (pss *PostgresStatusStore) getQueryRowFn(ctx context.Context) (queryRowFn queryRowFn, isTxSet bool) {
	tx := getTxFromContext(ctx)
	if tx == nil {
		queryRowFn = pss.db.QueryRow
		isTxSet = false
		return
	}

	queryRowFn = tx.QueryRow
	isTxSet = true
	return
}

// transactorKey is the unexported context key used to store an active pgx.Tx.
var transactorKey struct{}

// createContextWithTx returns a new context carrying the given transaction.
func createContextWithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, transactorKey, tx)
}

// getTxFromContext retrieves the pgx.Tx from ctx, or returns nil if none is set.
func getTxFromContext(ctx context.Context) pgx.Tx {
	value := ctx.Value(transactorKey)
	if value == nil {
		return nil
	}
	return value.(pgx.Tx)
}
