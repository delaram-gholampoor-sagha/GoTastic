package mysql

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"git.ice.global/packages/beeorm/v4"
	"github.com/delaram/GoTastic/internal/domain"
	"github.com/delaram/GoTastic/internal/repository"
)

type beeTx struct{ db *beeorm.DB }

func (t *beeTx) Commit(ctx context.Context) error   { t.db.Commit(); return nil }
func (t *beeTx) Rollback(ctx context.Context) error { t.db.Rollback(); return nil }

type OutboxRepo struct {
	engine *beeorm.Engine
}

func NewOutboxRepo(engine *beeorm.Engine) *OutboxRepo { return &OutboxRepo{engine: engine} }

func (r *OutboxRepo) BeginTx(ctx context.Context) (repository.Tx, error) {
	db := r.engine.GetMysql()
	db.Begin()
	return &beeTx{db: db}, nil
}

func (r *OutboxRepo) Insert(_ context.Context, _ repository.Tx, msg repository.OutboxMessage) error {
	var headersPtr *string
	if len(msg.Headers) > 0 {
		b, _ := json.Marshal(msg.Headers)
		s := string(b)
		headersPtr = &s
	}

	e := &domain.Outbox{
		AggregateType: msg.AggregateType,
		AggregateID:   msg.AggregateID,
		EventType:     msg.EventType,
		Payload:       msg.Payload,
		Headers:       headersPtr,
		Status:        "pending",
		Attempts:      0,
		AvailableAt:   time.Now().UTC(),
	}

	fl := r.engine.NewFlusher()
	fl.Track(e)
	return fl.FlushWithCheck()
}

func (r *OutboxRepo) FetchAndLock(ctx context.Context, limit int, lockForSeconds int) ([]repository.LockedOutboxRow, error) {
	tx, err := r.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	db := tx.(*beeTx).db
	now := time.Now()
	db.Exec(`
        UPDATE outbox
        SET status = 'pending'
        WHERE status = 'pending'
          AND available_at <= ?
        ORDER BY id
        LIMIT ?
    `, now, limit)
	rows, close := db.Query(`
    SELECT id, aggregate_type, aggregate_id, event_type, payload, attempts
    FROM outbox
    WHERE status = 'pending'
      AND available_at <= ?
    ORDER BY id
`, now)
	defer close()

	var out []repository.LockedOutboxRow
	for rows.Next() {
		var (
			id        uint64
			aggType   string
			aggID     string
			eventType string
			payload   []byte
			attempts  int
		)
		var scanErr error
		func() {
			defer func() {
				if r := recover(); r != nil {
					scanErr = fmt.Errorf("panic in rows.Scan: %v", r)
				}
			}()
			rows.Scan(&id, &aggType, &aggID, &eventType, &payload, &attempts)
		}()
		if scanErr != nil {
			log.Printf("Failed to scan row: %v", scanErr)
			return nil, scanErr
		}
		log.Printf("Scanned row: id=%d, aggType=%s, aggID=%s, eventType=%s, attempts=%d", id, aggType, aggID, eventType, attempts)
		out = append(out, repository.LockedOutboxRow{
			ID:            id,
			AggregateType: aggType,
			AggregateID:   aggID,
			EventType:     eventType,
			Payload:       payload,
			Attempts:      attempts,
		})
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return out, nil
}
func (r *OutboxRepo) MarkPublished(ctx context.Context, id uint64) error {
	var row domain.Outbox
	if ok := r.engine.LoadByID(id, &row); !ok {
		return repository.ErrNotFound
	}

	row.Status = "published"

	fl := r.engine.NewFlusher()
	fl.Track(&row)
	return fl.FlushWithCheck()
}

func (r *OutboxRepo) MarkFailed(ctx context.Context, id uint64, nextAvailableAt string, errMsg string) error {
	var row domain.Outbox
	if ok := r.engine.LoadByID(id, &row); !ok {
		return repository.ErrNotFound
	}

	ts, err := time.Parse(time.RFC3339, nextAvailableAt)
	if err != nil {
		if t2, err2 := time.Parse("2006-01-02 15:04:05", nextAvailableAt); err2 == nil {
			ts = t2
		} else {
			ts = time.Now().Add(time.Minute)
		}
	}

	row.Status = "pending"
	row.Attempts = row.Attempts + 1
	row.AvailableAt = ts

	fl := r.engine.NewFlusher()
	fl.Track(&row)
	return fl.FlushWithCheck()
}

func (r *OutboxRepo) GetByID(ctx context.Context, id uint64) (*domain.Outbox, error) {
	var row domain.Outbox
	where := beeorm.NewWhere("ID = ?", id)
	if ok := r.engine.SearchOne(where, &row); !ok {
		return nil, repository.ErrNotFound
	}
	return &row, nil
}

func (r *OutboxRepo) Delete(ctx context.Context, id uint64) error {
	var row domain.Outbox
	if ok := r.engine.LoadByID(id, &row); !ok {
		return repository.ErrNotFound
	}
	fl := r.engine.NewFlusher()
	fl.Delete(&row)
	return fl.FlushWithCheck()
}

func truncateErr(s string) string {
	if len(s) > 2000 {
		return s[:2000]
	}
	return s
}
