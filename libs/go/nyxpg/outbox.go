package nyxpg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// OutboxRecord is one pending row of the transactional outbox.
type OutboxRecord struct {
	ID           int64
	Tenant       string
	EventType    string
	PartitionKey string
	Payload      []byte
}

// Outbox implements the transactional outbox pattern (RFC-012 O.4): a
// domain write and its event enqueue commit atomically in the same PG
// transaction; libs/go/nyxbus's relay goroutine polls Pending and calls
// MarkPublished once each row lands on Kafka.
//
// Expects a table shaped like (created by the owning service's T-004
// migration, not by this library):
//
//	CREATE TABLE outbox_events (
//	  id            BIGSERIAL PRIMARY KEY,
//	  tenant_id     UUID NOT NULL DEFAULT current_setting('nydux.tenant')::uuid,
//	  event_type    TEXT NOT NULL,
//	  partition_key TEXT NOT NULL,
//	  payload       BYTEA NOT NULL,
//	  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
//	  published_at  TIMESTAMPTZ
//	);
//
// tenant_id defaults from the GUC WithTenantTx already set for this
// transaction, so Enqueue itself takes no tenant argument — it must only
// ever be called from within a WithTenantTx callback.
type Outbox struct {
	pool *Pool
}

// NewOutbox constructs an Outbox bound to pool.
func NewOutbox(pool *Pool) *Outbox { return &Outbox{pool: pool} }

// Enqueue records an event for relay within tx — callers run this in the
// same transaction as the domain write it accompanies, inside a
// WithTenantTx callback (tenant_id is populated from the GUC by column
// default, not passed here).
func (o *Outbox) Enqueue(ctx context.Context, tx pgx.Tx, eventType, partitionKey string, payload []byte) error {
	const q = `INSERT INTO outbox_events (event_type, partition_key, payload) VALUES ($1,$2,$3)`
	if _, err := tx.Exec(ctx, q, eventType, partitionKey, payload); err != nil {
		return fmt.Errorf("nyxpg: outbox enqueue: %w", err)
	}
	return nil
}

// Pending returns up to limit unpublished rows, oldest first.
func (o *Outbox) Pending(ctx context.Context, limit int) ([]OutboxRecord, error) {
	const q = `SELECT id, tenant_id, event_type, partition_key, payload
	           FROM outbox_events WHERE published_at IS NULL ORDER BY id ASC LIMIT $1`
	rows, err := o.pool.Raw().Query(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("nyxpg: outbox pending: %w", err)
	}
	defer rows.Close()

	var out []OutboxRecord
	for rows.Next() {
		var r OutboxRecord
		if err := rows.Scan(&r.ID, &r.Tenant, &r.EventType, &r.PartitionKey, &r.Payload); err != nil {
			return nil, fmt.Errorf("nyxpg: outbox scan: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// MarkPublished marks ids as relayed. No-op for an empty slice.
func (o *Outbox) MarkPublished(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	const q = `UPDATE outbox_events SET published_at = now() WHERE id = ANY($1)`
	if _, err := o.pool.Raw().Exec(ctx, q, ids); err != nil {
		return fmt.Errorf("nyxpg: outbox mark published: %w", err)
	}
	return nil
}
