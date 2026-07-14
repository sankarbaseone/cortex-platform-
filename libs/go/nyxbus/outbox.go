package nyxbus

import (
	"context"
	"fmt"
	"time"

	commonv1 "github.com/nydux/platform/api/nydux/common/v1"
	"github.com/nydux/platform/libs/go/nyxpg"
)

// OutboxReader is implemented by libs/go/nyxpg.Outbox.
type OutboxReader interface {
	Pending(ctx context.Context, limit int) ([]nyxpg.OutboxRecord, error)
	MarkPublished(ctx context.Context, ids []int64) error
}

// envelopePublisher is the narrow slice of *Producer that OutboxPublisher
// needs; kept as an interface (rather than depending on *Producer directly)
// so the relay logic is testable without a live Kafka broker.
type envelopePublisher interface {
	Publish(ctx context.Context, topic string, env *commonv1.EventEnvelope) error
}

// OutboxPublisher relays rows from an OutboxReader to Kafka, giving a
// domain write and its event publication atomic commit (RFC-012 O.4).
type OutboxPublisher struct {
	reader       OutboxReader
	producer     envelopePublisher
	pollInterval time.Duration
	batchSize    int
}

// Option configures an OutboxPublisher.
type Option func(*OutboxPublisher)

// WithPollInterval overrides the default 250ms poll interval.
func WithPollInterval(d time.Duration) Option { return func(o *OutboxPublisher) { o.pollInterval = d } }

// WithBatchSize overrides the default batch size of 500.
func WithBatchSize(n int) Option { return func(o *OutboxPublisher) { o.batchSize = n } }

// NewOutboxPublisher constructs a relay polling reader every 250ms in
// batches of 500 by default (ECD-004 §4.2 exemplar main.go comment).
func NewOutboxPublisher(reader OutboxReader, producer *Producer, opts ...Option) *OutboxPublisher {
	o := &OutboxPublisher{
		reader:       reader,
		producer:     producer,
		pollInterval: 250 * time.Millisecond,
		batchSize:    500,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// Relay returns the nyxrun.Runnable-shaped goroutine body: `Run(ctx) error`.
func (o *OutboxPublisher) Relay() *Relay { return &Relay{o: o} }

// Relay is the running form of an OutboxPublisher, passed to nyxrun.Group.
type Relay struct{ o *OutboxPublisher }

// Run polls the outbox on pollInterval until ctx is canceled. Transient
// errors are swallowed (crash-only design: the next tick retries) rather
// than killing the relay goroutine.
func (r *Relay) Run(ctx context.Context) error {
	ticker := time.NewTicker(r.o.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			_ = r.o.relayOnce(ctx)
		}
	}
}

func (o *OutboxPublisher) relayOnce(ctx context.Context) error {
	rows, err := o.reader.Pending(ctx, o.batchSize)
	if err != nil {
		return fmt.Errorf("nyxbus: outbox relay: pending: %w", err)
	}
	if len(rows) == 0 {
		return nil
	}
	published := make([]int64, 0, len(rows))
	for _, row := range rows {
		env := &commonv1.EventEnvelope{
			TenantId:     row.Tenant,
			Type:         row.EventType,
			Payload:      row.Payload,
			PartitionKey: row.PartitionKey,
		}
		if err := o.producer.Publish(ctx, row.EventType, env); err != nil {
			// Stop at first failure so ordering/at-least-once holds for the
			// remaining unpublished rows; next tick resumes from here.
			break
		}
		published = append(published, row.ID)
	}
	if len(published) == 0 {
		return nil
	}
	if err := o.reader.MarkPublished(ctx, published); err != nil {
		return fmt.Errorf("nyxbus: outbox relay: mark published: %w", err)
	}
	return nil
}
