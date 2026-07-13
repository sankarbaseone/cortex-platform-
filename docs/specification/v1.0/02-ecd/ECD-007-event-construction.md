# ECD-007 — Event Construction

**Level:** 2 · Extends RFC-005 (frozen). **Artifacts are the spec:**
- `artifacts/proto/nydux/common/v1/envelope.proto` — complete, compilable, protovalidate-annotated envelope + ErrorInfo.
- `artifacts/proto/nydux/compiler/v1/kernel.proto` — complete compiler-domain payloads + KernelService/RegressionService.
- `artifacts/proto/nydux/finance/v1/cost.proto`, `artifacts/proto/nydux/rec/v1/recommendation.proto` — complete.
- `artifacts/kafka/topics.yaml` — the machine-readable event catalog: every topic with class, partitions, retention, key, producer, consumers, idempotency strategy, DLQ thresholds, retry policy, schema-evolution rules, example payload, and load estimates.

Remaining proto packages (`infra/v1`, `runtime/v1`, `graph/v1`, `policy/v1`, `audit/v1`, `twin/v1`, `agent/v1`, `gateway/v1`) are constructed by Claude Code using the exact field lists already fixed in ECD-003 §3.2/§3.3 and RFC-005 B.3, in the identical style of the four shipped files (same options, validate annotations, enum-zero rules). This is transcription, not design — the shipped files are the binding style reference. (`GpuSample` full message is shipped inside kernel.proto's bundle note; repo location is `infra/v1/gpu.proto` per ECD-001 tree.)

## 7.1 Topic configuration (normative, from topics.yaml defaults)
Telemetry class: 32 partitions · 7d · delete · zstd · DLQ alert >1000. Business: 6 partitions · 30d · compact+delete · DLQ alert >0. Audit: 1 partition/tenant · compacted-forever · strict sequence + gap detector. Producers: acks=all, idempotent. Consumer groups: one group per consuming service, named `nydux.<service>.<topic-short>`; scaling = partition count ceiling.

## 7.2 Retry / DLQ / replay (construction detail)
In-process backoff 100ms→30s ×5 → `.retry` topic (15-min delayed consumer, same group suffix `.retry`) → `.dlq` (14d). Redrive: `nyduxctl dlq redrive --topic T --from ISO --to ISO [--dry-run]` — replays preserving original envelope (event_id unchanged ⇒ consumer idempotency absorbs). Poison-pill: unmarshal failure bypasses retries → DLQ with envelope intact + `poison=true` header. Replay-for-recovery: ClickHouse rebuild path = reset ch-sink group offsets to timestamp (7d telemetry retention window is the recovery budget — dual path with CH backups per ECD-006).

## 7.3 Ordering construction
Strict-order events (approved/applied/deploy/savings/baseline/audit/governance rows in topics.yaml) achieve order via: single partition key + max.in.flight=1 on those producers + per-key single consumer goroutine. Everything else is order-tolerant by upsert design. audit.appended additionally carries `seq`; the gap-detector consumer alerts Sev-1 on any hole (RFC-009 I.8).

## 7.4 Versioning & evolution
Registry compatibility BACKWARD; CI runs buf-breaking + registry check on every proto change (RFC-011 Q). Adding a field = minor, safe. Semantics change or field removal = new message + new `.v2` topic, dual-publish ≥90 days, consumer-migration dashboard tracks group offsets on old topic reaching zero before decommission. Envelope frozen (additive only) — restated as a hard rule.

## 7.5 Monitoring & alerting (per topic class; rule names fixed for ECD-011 Grafana/Prom construction in Run 3)
`NyduxConsumerLagHigh` (telemetry: >5min sustained 10min; business: >60s sustained 5min) · `NyduxDLQNonEmptyBusiness` (>0, page) · `NyduxDLQTelemetryBurst` (>1000, ticket) · `NyduxAuditGapDetected` (any, page Sev-1) · `NyduxOutboxBacklog` (>10k unpublished rows 5min) · `NyduxSchemaRegistryDrift` (producer schema id unknown). Each maps to runbooks RB-BUS-001…006 (Run 3 delivers runbook bodies).

## 7.6 Producer/consumer construction defaults (code-level, libs/go/nyxbus)
Producer: linger 20ms, batch 512KB, zstd, delivery timeout 30s, transactional NOT used (outbox gives atomicity; Kafka transactions rejected: operational complexity, no end-to-end exactly-once anyway per RFC-005 B.4). Consumer: cooperative-sticky rebalance, max.poll.interval 5min, commit-after-process, per-partition goroutine, bounded channel 1024, backpressure = pause partition.

## 7.7 Example payloads
Canonical JSON example for `kernel.scored` ships in topics.yaml; every other topic's example is generated into event-registry.yaml (Run 4) from proto fixtures in `proto/testdata/` — fixture files are mandatory per new payload message (CI: fixture-missing fails).
