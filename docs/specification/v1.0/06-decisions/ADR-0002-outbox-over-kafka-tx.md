# ADR-0002: Transactional outbox instead of Kafka transactions
Status: accepted · Date: 2026-07-13 · Owner: platform
## Context
DB write + event publish must be atomic (ECD-007 §7.6 decision, recorded as ADR).
## Decision
Per-service outbox table in the same PG transaction; relay goroutine polls 250ms/batch 500; consumers idempotent by event_id/natural key.
## Alternatives rejected
Kafka transactions (operational complexity, still no end-to-end exactly-once across DB+consumers); dual-write with retries (loss window); CDC/Debezium (new infra component, ordering complexity).
## Consequences
At-least-once delivery with idempotent effect; NyduxOutboxBacklog alert guards relay health.
## Compliance
Implements RFC-005 B.4/RFC-012 O.4.
