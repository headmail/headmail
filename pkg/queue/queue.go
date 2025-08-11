package queue

import (
	"context"
	"encoding/json"
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusReserved Status = "reserved"
	StatusDone     Status = "done"
	StatusFailed   Status = "failed"
)

// QueueItem represents a generic queue entry.
// Payload is opaque bytes (typically JSON) and interpreted by the consumer.
type QueueItem struct {
	ID         string          `json:"id"`                    // UUID
	Type       string          `json:"type"`                  // arbitrary topic/type
	Payload    json.RawMessage `json:"payload"`               // opaque payload (JSON)
	UniqueKey  *string         `json:"unique_key,omitempty"`  // optional deduplication key
	Status     Status          `json:"status"`                // pending/reserved/done/failed
	ReservedBy *string         `json:"reserved_by,omitempty"` // worker id that reserved the item
	ReservedAt *int64          `json:"reserved_at,omitempty"`
	CreatedAt  int64           `json:"created_at"`
}

// Queue is an abstract queue interface. All methods accept context so that
// transaction context (tx) can be passed via context values.
type Queue interface {
	// Enqueue inserts a new item into the queue. If the implementation supports
	// deduplication via UniqueKey, duplicate inserts should be ignored.
	Enqueue(ctx context.Context, item *QueueItem) error

	// Claim atomically reserves up to `limit` pending items that are ready
	// Claimed items should have Status changed to "reserved" and reserved_by set.
	Claim(ctx context.Context, workerID string, limit int) ([]*QueueItem, error)

	// Ack marks the queue item as successfully processed (can delete or mark done).
	Ack(ctx context.Context, id string) error

	// Fail marks the queue item as failed for this attempt
	Fail(ctx context.Context, id string, reason string) error
}

type Handler = func(txCtx context.Context, workerID string, item *QueueItem) error
