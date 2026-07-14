package nyxbus

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"

	commonv1 "github.com/nydux/platform/api/nydux/common/v1"
)

// Producer publishes EventEnvelope-wrapped events to Kafka with an
// idempotent, acks=all producer (RFC-005 B.4: "Producers: acks=all,
// idempotent producer on").
type Producer struct {
	client *kgo.Client
}

// NewProducer opens a Kafka producer client. Idempotency is enabled by
// default in franz-go (disabled only via kgo.DisableIdempotentWrite, which
// we never set).
func NewProducer(cfg Config) (*Producer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ClientID(cfg.ClientID),
		kgo.RequiredAcks(kgo.AllISRAcks()),
	)
	if err != nil {
		return nil, fmt.Errorf("nyxbus: new producer: %w", err)
	}
	return &Producer{client: client}, nil
}

// Close releases the underlying Kafka client.
func (p *Producer) Close() { p.client.Close() }

// Publish synchronously produces env to topic, keyed by env.PartitionKey
// (RFC-005 B.2/B.3: partition key is documented per event in the catalog).
func (p *Producer) Publish(ctx context.Context, topic string, env *commonv1.EventEnvelope) error {
	payload, err := proto.Marshal(env)
	if err != nil {
		return fmt.Errorf("nyxbus: marshal envelope: %w", err)
	}
	rec := &kgo.Record{Topic: topic, Key: []byte(env.GetPartitionKey()), Value: payload}
	result := p.client.ProduceSync(ctx, rec)
	if err := result.FirstErr(); err != nil {
		return fmt.Errorf("nyxbus: produce %s: %w", topic, err)
	}
	return nil
}

// publishRaw produces a pre-encoded value with the given key, used by the
// retry/DLQ paths which forward the original wire bytes unmodified
// (RFC-005 B.4: "poison-pill guard ... DLQ with envelope intact").
func (p *Producer) publishRaw(ctx context.Context, topic string, key, value []byte) error {
	rec := &kgo.Record{Topic: topic, Key: key, Value: value}
	result := p.client.ProduceSync(ctx, rec)
	if err := result.FirstErr(); err != nil {
		return fmt.Errorf("nyxbus: produce %s: %w", topic, err)
	}
	return nil
}
