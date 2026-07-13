# RFC-005 — Event Bus & Telemetry

**Status:** Approved · **Extends:** V1 Phase 4.1, 5.4 · **Owns Section B** entirely.

## B.1 Buses
- **Kafka (CP):** durable domain events + telemetry ingest. Redpanda drop-in allowed (OQ-04).
- **NATS (DP-local + CP control):** low-latency agent/control messaging, request-reply. No durable business facts on NATS.
Contract: every domain fact that other services rely on MUST be a Kafka event with a registered schema.

## B.2 Envelope (protobuf, registered)
```proto
message EventEnvelope {
  string event_id = 1;        // UUIDv7 — idempotency key
  string tenant_id = 2;
  string cluster_id = 3;      // "" for CP-origin
  string type = 4;            // e.g. "compiler.kernel.scored"
  uint32 schema_version = 5;  // payload schema semver-major
  int64  event_time_ns = 6;   // occurrence (device clock, NTP-disciplined)
  int64  ingest_time_ns = 7;
  string trace_id = 8;        // W3C traceparent
  string producer = 9;        // service@version
  bytes  payload = 10;        // typed message per catalog
  string partition_key = 11;  // documented per event
}
```

## B.3 Event Catalog (complete; topic = `nydux.<name>.v1`)
Columns: **Event · Producer → Consumers · Key payload fields · Partition key · Ordering need**

**Telemetry**
- `infra.gpu.sample` · collector → CH-sink, anomaly · dcgm fields, gpu_uuid, pod refs · (cluster,gpu_uuid) · per-GPU order
- `infra.node.health` · collector → infra-svc, failure-pred · xid, ecc, thermal · (cluster,node) · per-node
- `runtime.serving.metrics` · runtime-analyzer → CH-sink, finance · ttft, tpot, batch, kv_hit, req counts · (cluster,deploy) · per-deploy
- `runtime.nccl.collective` · collector → runtime-analyzer · algo, bytes, dur, ranks · (cluster,job) · per-job
- `runtime.trace.captured` · runtime-analyzer → blob-ref only · trace_uri(in-tenant), summary · (cluster,job) · none

**Compiler pipeline (Section B required examples all present)**
- `compiler.model.loaded` · SDK hook → analyzer · model_id, fw_ver · (tenant,model) · none
- `compiler.graph.captured` · SDK/CI → analyzer · fx/hlo ref(in-tenant), op_stats · (tenant,model) · none
- `compiler.kernel.generated` · analyzer → registry · family_hash, source(triton/inductor/xla), ir_features · (tenant,family) · per-family
- `compiler.kernel.compiled` · analyzer → registry · kernel_hash, toolchain, flags · (tenant,kernel) · per-kernel
- `compiler.ptx.generated` · analyzer → registry · kernel_hash, ptx_features(regs,instr mix) · (tenant,kernel) · per-kernel
- `compiler.sass.generated` · analyzer → registry · kernel_hash, sass_features(spills,occ inputs), decode_confidence · (tenant,kernel) · per-kernel
- `compiler.kernel.scored` · analyzer → graph, recommender, UI-feed · kes, components, kes_model_version, status · (tenant,kernel) · per-kernel
- `compiler.regression.detected` · regression-svc → graph, notify, CI-gate · family, from/to toolchain, Δperf, cri_contrib, mechanism_tags · (tenant,family) · per-family
- `compiler.benchmark.completed` · bench-runner → regression-svc, graph · run stats, env fingerprint · (tenant,run) · none

**Runtime execution**
- `runtime.job.submitted|started|finished` · runtime-analyzer → finance, twin · job spec, P config, duration · (tenant,job) · per-job strict
- `infra.gpu.scheduled|started|released` · collector(k8s watch) → finance · pod↔gpu binding, mig profile · (cluster,gpu_uuid) · per-GPU
- `runtime.inference.started|finished` · sampled · req class, tokens{in,out} · (cluster,deploy) · per-deploy

**Business**
- `finance.cost.calculated` · finance-svc → graph, UI · slice_id, usd, basis, window · (tenant,slice) · per-slice
- `finance.savings.reported` · savings-svc → UI, audit · period, S, method refs, baseline_ver · (tenant,period) · strict per-tenant
- `finance.baseline.reanchored` · savings-svc → audit · reason, cosign refs · (tenant) · strict

**Recommendations & governance**
- `rec.created` · recommender → UI, policy · rec fields (RFC-002 §2.7) · (tenant,rec) · per-rec
- `rec.approved|rejected` · control-plane-api(user action) → agent-orch, audit · actor, rationale · (tenant,rec) · per-rec strict
- `rec.applied` · agent-orch → verify-runner, audit · change ref, rollback token · (tenant,rec) · per-rec strict
- `rec.verified` · verify-runner → graph(RESULTED_IN), savings · measured gain · (tenant,rec) · per-rec
- `policy.decision` · policy-svc → audit, notify · policy_id, verdict, subject · (tenant,subject) · per-subject
- `governance.toolchain.approved|revoked` · policy-svc → CI-gate, notify · toolchain, scope · (tenant) · strict
- `deploy.triggered|completed|rolledback` · agent-orch → audit, twin(calibration) · change set, env · (tenant,change) · strict per-change
- `audit.appended` · audit-svc → (ledger only) · hash-chain entry · (tenant) · STRICT single-partition per tenant

**Agents & system**
- `agent.task.created|completed|failed` · agent-orch ↔ agents (NATS mirror→Kafka) · task, tool calls summary, tokens used · (tenant,task) · per-task
- `system.schema.registered`, `system.tenant.provisioned|offboarded`, `notify.dispatched`.

## B.4 Retry / DLQ / Idempotency (uniform policy)
- Producers: acks=all, idempotent producer on, exactly-once NOT assumed end-to-end; consumers MUST be idempotent by `event_id` (dedup table or upsert-by-natural-key — each consumer documents which in RFC-014 entry).
- Consumer retry: in-process exp backoff (100ms→30s, jitter, max 5) → per-topic retry topic (`.retry`, 15-min delayed consumer) → DLQ (`.dlq`, 14-day retention, alert at >0 for business topics, >1k for telemetry).
- DLQ redrive tool: `nyduxctl dlq redrive --topic --from --to` with dry-run.
- Poison-pill guard: payload parse failure goes straight to DLQ with envelope intact.

## B.5 Ordering
Only keys marked "strict" require order; achieved by single partition-key + single consumer per key group. Everything else tolerates reordering by design (upserts keyed by natural key + event_time). audit.appended is single-partition-per-tenant with sequence numbers; a gap detector runs continuously (RFC-009 §I.8).

## B.6 Versioning
Confluent-style schema registry; BACKWARD compatibility enforced in CI. Breaking change ⇒ new topic `.v2`, dual-publish window ≥90 days, consumer migration tracked in a registry dashboard. Envelope itself is frozen (only additive).

## B.7 Backfill & clock discipline
Collectors spool locally on link loss (RFC-001 A.6) and replay with ORIGINAL event_time; all analytics window on event_time with 6h lateness watermark (ClickHouse ingestion handles late rows; Timescale continuous aggregates use 6h refresh lag). NTP required; collector rejects start if clock skew >2s and reports it.

## B.8 Throughput & partitions (initial)
Telemetry topics: 32 partitions, 7d retention, zstd. Business topics: 6 partitions, 30d, zstd, cleanup=compact+delete where natural-keyed. Budget: 100k GPU fleet ≈ 10k samples/s ≈ trivially within a 3-broker cluster; headroom 20×.

## B.9 Testing
Contract tests generated from registry (producer fixtures replayed against consumer test harness); chaos test: broker kill during ingest must lose zero acked events; backfill test: 24h spool replay correctness on watermark boundaries.
