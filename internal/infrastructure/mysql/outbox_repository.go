// internal/infrastructure/mysql/outbox_repository.go
package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/delaram/GoTastic/internal/repository"
	"github.com/google/uuid"
)

type sqlTx struct{ *sql.Tx }

func (t *sqlTx) Commit(ctx context.Context) error   { return t.Tx.Commit() }
func (t *sqlTx) Rollback(ctx context.Context) error { return t.Tx.Rollback() }

type OutboxRepo struct {
	db *sql.DB
}

func NewOutboxRepo(db *sql.DB) *OutboxRepo { return &OutboxRepo{db: db} }

func (r *OutboxRepo) BeginTx(ctx context.Context) (repository.Tx, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	return &sqlTx{tx}, nil
}

func (r *OutboxRepo) Insert(ctx context.Context, tx repository.Tx, msg repository.OutboxMessage) error {
	sqltx := tx.(*sqlTx).Tx

	var headersJSON *string
	if msg.Headers != nil {
		b, _ := json.Marshal(msg.Headers)
		s := string(b)
		headersJSON = &s
	}

	_, err := sqltx.ExecContext(ctx, `
        INSERT INTO outbox (aggregate_type, aggregate_id, event_type, payload, headers)
        VALUES (?, ?, ?, JSON_EXTRACT(?, '$'), ?)
    `, msg.AggregateType, msg.AggregateID, msg.EventType, string(msg.Payload), headersJSON)
	return err
}

func (r *OutboxRepo) FetchAndLock(ctx context.Context, limit int, lockForSeconds int) ([]repository.LockedOutboxRow, error) {
	lockToken := uuid.New().String()
	now := time.Now()
	lockUntil := now.Add(time.Duration(lockForSeconds) * time.Second)

	// MySQL-safe: UPDATE ... ORDER BY ... LIMIT (no subquery-in-IN)
	_, err := r.db.ExecContext(ctx, `
		UPDATE outbox
		SET lock_token = ?, locked_until = ?, status = 'pending'
		WHERE status = 'pending'
		  AND available_at <= ?
		  AND (locked_until IS NULL OR locked_until < ?)
		ORDER BY id
		LIMIT ?;
	`, lockToken, lockUntil, now, now, limit)
	if err != nil {
		return nil, err
	}

	// Read back the rows we just locked
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, aggregate_type, aggregate_id, event_type, payload, attempts
		FROM outbox
		WHERE lock_token = ? AND locked_until >= ?
		ORDER BY id;
	`, lockToken, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []repository.LockedOutboxRow
	for rows.Next() {
		var rID uint64
		var aggType, aggID, evt string
		var payload []byte
		var attempts int
		if err := rows.Scan(&rID, &aggType, &aggID, &evt, &payload, &attempts); err != nil {
			return nil, err
		}
		out = append(out, repository.LockedOutboxRow{
			ID: rID, AggregateType: aggType, AggregateID: aggID, EventType: evt, Payload: payload, Attempts: attempts,
		})
	}
	return out, rows.Err()
}

func (r *OutboxRepo) MarkPublished(ctx context.Context, id uint64) error {
	_, err := r.db.ExecContext(ctx, `
        UPDATE outbox
        SET status = 'published', published_at = NOW(), lock_token = NULL, locked_until = NULL
        WHERE id = ?
    `, id)
	return err
}

func (r *OutboxRepo) MarkFailed(ctx context.Context, id uint64, nextAvailableAt string, errMsg string) error {
	_, err := r.db.ExecContext(ctx, `
        UPDATE outbox
        SET status = 'pending',
            attempts = attempts + 1,
            available_at = ?,
            error = ?,
            lock_token = NULL,
            locked_until = NULL
        WHERE id = ?
    `, nextAvailableAt, truncateErr(errMsg), id)
	return err
}

func truncateErr(s string) string {
	if len(s) > 2000 {
		return s[:2000]
	}
	return s
}
