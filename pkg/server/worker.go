package server

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/headmail/headmail/pkg/queue"
	"github.com/headmail/headmail/pkg/repository"
)

// Worker is a simple queue worker that processes generic queue items.
// For now it understands items of Type == "delivery" where Payload is {"delivery_id":"<id>"}.
// The mail send is a stub (logged).
type Worker struct {
	db       repository.DB
	q        queue.Queue
	handlers map[string]queue.Handler

	// claim limit per iteration
	limit int
	// sleep duration when queue empty or error
	idleSleep time.Duration
}

// NewWorker constructs a Worker.
func NewWorker(db repository.DB, q queue.Queue) *Worker {
	return &Worker{
		db:        db,
		q:         q,
		handlers:  make(map[string]queue.Handler),
		limit:     1,
		idleSleep: time.Second,
	}
}

func (w *Worker) SetHandler(name string, handler queue.Handler) error {
	w.handlers[name] = handler
	return nil
}

// Start runs the worker loop until ctx is cancelled.
func (w *Worker) Start(ctx context.Context, workerID string) {
	log.Printf("worker %s started", workerID)
	for {
		select {
		case <-ctx.Done():
			log.Printf("worker %s stopping: context done", workerID)
			return
		default:
		}

		items, err := w.q.Claim(ctx, workerID, w.limit)
		if err != nil {
			log.Printf("worker %s: claim error: %v", workerID, err)
			time.Sleep(w.idleSleep)
		}
		if len(items) == 0 {
			time.Sleep(w.idleSleep)
		}

		for _, it := range items {
			// process each item, do not block the whole worker loop on one failing item
			if err := w.processItem(ctx, workerID, it); err != nil {
				log.Printf("worker %s: failed processing item %s: %v", workerID, it.ID, err)
			}
		}
	}
}

func (w *Worker) processItem(ctx context.Context, workerID string, it *queue.QueueItem) error {
	handler, ok := w.handlers[it.Type]
	if !ok {
		// No handler registered: treat as permanent failure (ack to remove).
		return fmt.Errorf("no handler for '%s'", it.Type)
	}

	// Start DB transaction so handler can update domain and we can ack atomically.
	txCtx, err := w.db.Begin(ctx)
	if err != nil {
		// If we cannot start a transaction, mark item for retry.
		_ = w.q.Fail(ctx, it.ID, err.Error())
		return err
	}

	// Execute handler with transactional context.
	if err := handler(txCtx, workerID, it); err != nil {
		_ = w.db.Rollback(txCtx)
		_ = w.q.Fail(ctx, it.ID, err.Error())
		return err
	}

	// Ack the queue item within the same transaction and commit.
	if err := w.q.Ack(txCtx, it.ID); err != nil {
		_ = w.db.Rollback(txCtx)
		_ = w.q.Fail(ctx, it.ID, err.Error())
		return err
	}

	if err := w.db.Commit(txCtx); err != nil {
		_ = w.q.Fail(ctx, it.ID, err.Error())
		return err
	}

	return nil
}
