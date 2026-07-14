package nyxbus

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"

	commonv1 "github.com/nydux/platform/api/nydux/common/v1"
)

// Handler processes one decoded event. Return Poison(err) to signal a
// payload the handler cannot process (schema mismatch, corrupt bytes) —
// this skips retry and forwards straight to DLQ. Any other error triggers
// the in-process backoff/retry policy (RFC-005 B.4).
type Handler func(ctx context.Context, env *commonv1.EventEnvelope) error

// ConsumerConfig configures one topic's consumer loop.
type ConsumerConfig struct {
	Config
	Topic       string
	GroupID     string
	MaxRetries  int           // in-process retry attempts before handing off to the .retry topic; default 5 (RFC-005 B.4)
	BaseBackoff time.Duration // default 100ms (RFC-005 B.4)
	MaxBackoff  time.Duration // default 30s (RFC-005 B.4)
}

func (c *ConsumerConfig) setDefaults() {
	if c.MaxRetries <= 0 {
		c.MaxRetries = 5
	}
	if c.BaseBackoff <= 0 {
		c.BaseBackoff = 100 * time.Millisecond
	}
	if c.MaxBackoff <= 0 {
		c.MaxBackoff = 30 * time.Second
	}
}

// Consumer runs Handler over one Kafka topic with the platform's uniform
// retry/DLQ policy (RFC-005 B.4):
//   - envelope undecodable                 -> straight to <topic>.dlq, raw bytes intact
//   - Handler returns Poison(err)           -> straight to <topic>.dlq
//   - Handler returns any other error       -> in-process exp backoff
//     (100ms->30s, jittered, up to MaxRetries), then handed to <topic>.retry
//
// Consumers MUST be idempotent by event_id (dedup table or upsert-by-
// natural-key) — that guarantee lives in Handler, not here.
// rawPublisher is the narrow slice of *Producer the retry/DLQ paths need;
// an interface so processRecord is testable without a live Kafka broker.
type rawPublisher interface {
	publishRaw(ctx context.Context, topic string, key, value []byte) error
}

type Consumer struct {
	cfg     ConsumerConfig
	client  *kgo.Client
	dlq     rawPublisher
	handler Handler
}

// NewConsumer opens a Kafka consumer-group client for cfg.Topic and wires
// dlq (used to forward poison/exhausted-retry records to <topic>.dlq and
// <topic>.retry).
func NewConsumer(cfg ConsumerConfig, dlq *Producer, handler Handler) (*Consumer, error) {
	cfg.setDefaults()
	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ClientID(cfg.ClientID),
		kgo.ConsumerGroup(cfg.GroupID),
		kgo.ConsumeTopics(cfg.Topic),
	)
	if err != nil {
		return nil, fmt.Errorf("nyxbus: new consumer: %w", err)
	}
	return &Consumer{cfg: cfg, client: client, dlq: dlq, handler: handler}, nil
}

// Close releases the underlying Kafka client.
func (c *Consumer) Close() { c.client.Close() }

// Run implements the nyxrun.Runnable shape: polls fetches until ctx is
// canceled (crash-only design — consumers resume from committed offsets on
// restart, RFC-014 G.0).
func (c *Consumer) Run(ctx context.Context) error {
	for {
		fetches := c.client.PollFetches(ctx)
		if ctx.Err() != nil {
			return nil
		}
		if errs := fetches.Errors(); len(errs) > 0 {
			return fmt.Errorf("nyxbus: fetch %s: %v", c.cfg.Topic, errs)
		}
		fetches.EachRecord(func(rec *kgo.Record) {
			c.processRecord(ctx, rec)
		})
	}
}

func (c *Consumer) processRecord(ctx context.Context, rec *kgo.Record) {
	var env commonv1.EventEnvelope
	if err := proto.Unmarshal(rec.Value, &env); err != nil {
		// Envelope itself is undecodable: forward the raw wire bytes as-is.
		_ = c.dlq.publishRaw(ctx, c.cfg.Topic+".dlq", rec.Key, rec.Value)
		return
	}

	backoff := c.cfg.BaseBackoff
	for attempt := 0; ; attempt++ {
		err := c.handler(ctx, &env)
		if err == nil {
			return
		}
		if IsPoison(err) {
			_ = c.dlq.publishRaw(ctx, c.cfg.Topic+".dlq", rec.Key, rec.Value)
			return
		}
		if attempt >= c.cfg.MaxRetries {
			// In-process retries exhausted; hand off to the delayed retry
			// topic (separate consumer redelivers after its retention/delay
			// window — out of scope for this library, see nyduxctl dlq redrive).
			_ = c.dlq.publishRaw(ctx, c.cfg.Topic+".retry", rec.Key, rec.Value)
			return
		}
		sleep := jittered(backoff)
		select {
		case <-ctx.Done():
			return
		case <-time.After(sleep):
		}
		backoff *= 2
		if backoff > c.cfg.MaxBackoff {
			backoff = c.cfg.MaxBackoff
		}
	}
}

func jittered(d time.Duration) time.Duration {
	if d <= 0 {
		return 0
	}
	return d/2 + time.Duration(rand.Int63n(int64(d)/2+1))
}
