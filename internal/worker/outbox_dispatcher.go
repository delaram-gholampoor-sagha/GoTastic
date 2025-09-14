// internal/worker/outbox_dispatcher.go
package worker

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"time"

	"github.com/delaram/GoTastic/internal/domain"
	"github.com/delaram/GoTastic/internal/repository"
)

type OutboxDispatcher struct {
	outbox repository.OutboxRepository
	// keep your existing publisher; we’ll adapt by event_type
	stream repository.StreamPublisher
	// config
	batchSize      int
	lockForSeconds int
	maxAttempts    int
}

func NewOutboxDispatcher(outbox repository.OutboxRepository, stream repository.StreamPublisher) *OutboxDispatcher {
	return &OutboxDispatcher{
		outbox: outbox, stream: stream,
		batchSize: 100, lockForSeconds: 30, maxAttempts: 10,
	}
}

func (d *OutboxDispatcher) Run(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.tick(ctx)
		}
	}
}

func (d *OutboxDispatcher) tick(ctx context.Context) {
	rows, err := d.outbox.FetchAndLock(ctx, d.batchSize, d.lockForSeconds)
	if err != nil {
		log.Printf("outbox fetch error: %v", err)
		return
	}
	for _, row := range rows {
		if err := d.handle(ctx, row); err != nil {
			// compute backoff and reschedule
			next := nextBackoff(row.Attempts+1, time.Second, 10*time.Minute)
			_ = d.outbox.MarkFailed(ctx, row.ID, next.Format("2006-01-02 15:04:05"), err.Error())
			continue
		}
		_ = d.outbox.MarkPublished(ctx, row.ID)
	}
}

func (d *OutboxDispatcher) handle(ctx context.Context, row repository.LockedOutboxRow) error {
	switch row.EventType {
	case "todo.created":
		var todo domain.TodoItem
		if err := json.Unmarshal(row.Payload, &todo); err != nil {
			return err
		}
		return d.stream.PublishTodoItem(ctx, &todo)

	// add more event types here:
	// case "file.deleted": ...
	default:
		// treat unknown events as success (or fail; your call)
		return nil
	}
}

func nextBackoff(attempt int, base, max time.Duration) time.Time {
	// exponential with jitter-less cap
	mult := math.Pow(2, float64(attempt-1))
	d := time.Duration(mult) * base
	if d > max {
		d = max
	}
	return time.Now().Add(d)
}
