// Package nyxbus wraps Kafka producer/consumer access with the platform's
// event-bus contract (RFC-005): durable domain events over Kafka, uniform
// retry/DLQ/idempotency policy (B.4), and the outbox relay that gives a
// domain write + its event atomic commit (RFC-012 O.4 outbox pattern).
package nyxbus

// Config is embedded (as Kafka) in each service's own config struct.
type Config struct {
	Brokers  []string `envconfig:"KAFKA_BROKERS" required:"true"`
	ClientID string   `envconfig:"KAFKA_CLIENT_ID" default:"nydux"`
}
