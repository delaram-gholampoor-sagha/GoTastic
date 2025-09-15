// internal/repository/outbox.go
package repository

import "context"

type OutboxMessage struct {
	AggregateType string            // "todo"
	AggregateID   string            // todo.ID
	EventType     string            // "todo.created"
	Payload       []byte            // JSON
	Headers       map[string]string // optional
}

type OutboxRepository interface {
	// Insert within the same DB transaction as your aggregate write.
	Insert(ctx context.Context, tx Tx, msg OutboxMessage) error

	// Worker: fetch a batch of due messages and lock them for N seconds to avoid duplicate workers.
	FetchAndLock(ctx context.Context, limit int, lockForSeconds int) ([]LockedOutboxRow, error)

	// Mark publish result.
	MarkPublished(ctx context.Context, id uint64) error
	MarkFailed(ctx context.Context, id uint64, nextAvailableAt string, errMsg string) error
}

type LockedOutboxRow struct {
	ID            uint64
	AggregateType string
	AggregateID   string
	EventType     string
	Payload       []byte
	Attempts      int
}
