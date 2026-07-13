# RFC-014 — Microservice Catalog

**Status:** Approved · **Owns Section G.** Uniform contract first; per-service deltas after. Layout per RFC-012 O.2.

## G.0 Uniform service contract (applies to every service; deviations must be listed in the delta table)
- **Health:** `/healthz` (liveness: process up), `/readyz` (readiness: deps ping w/ 2s budget), `/startupz` for slow-boot (analyzers).
- **Metrics:** Prometheus `/metrics`; mandatory: RED (rate/errors/duration per RPC), consumer lag, DLQ depth, dependency-call histograms; service-specific listed below.
- **Logs:** structured JSON {ts, level, svc, tenant, trace_id, msg, fields}; payload contents NEVER logged (lint rule).
- **Tracing:** OTel; W3C propagation; spans on every RPC, consumer batch, DB call.
- **Failure recovery:** crash-only design; consumers resume from committed offsets; outbox relays idempotent.
- **Circuit breakers:** all outbound sync calls wrapped (gobreaker): 5 consecutive failures → open 30s → half-open probe; fallback documented per call (usually degrade/cached).
- **Timeouts/retries:** client defaults 10s deadline, 3 retries idempotent-only, jittered backoff; NEVER retry non-idempotent without idempotency key.
- **Config:** env-only, typed struct, validated at boot, printed (secrets redacted) at startup.
- **Secrets:** via Vault/ESO mounts (RFC-009 I.4).
- **Feature flags:** OpenFeature client; flags named `svc.feature`; default-off.
- **Resource limits:** requests=limits for memory (no OOM surprises), CPU requests set from load tests; HPA on primary saturation signal listed below.
- **Deployment:** Deployment ×≥3 (CP), PDB, topology spread, canary per RFC-011 J.6.
- **Security:** non-root, read-only rootfs, seccomp default, NetworkPolicy least-privilege.

## G.1 Service table (name · responsibility · API · owns store · scale signal · key metrics · notable failure handling)
| Service | Responsibility | API | Store | HPA signal | Extra metrics | Notable failure handling |
|---|---|---|---|---|---|---|
| edge-gateway (DP, Rust) | authn'd egress, privacy filter (D.9), spool | gRPC client-stream up | local disk ring | outbound queue depth | spool bytes, filter rejects | link-loss spool 24h; filter reject = hard error + alert (never "best effort" egress) |
| collector (DP, Rust) | DCGM/NVML/k8s-watch sampling | NATS local → gateway | none | n/a (DaemonSet) | sample lag, clock skew | skew>2s refuse start; DCGM absent → degraded fields flagged |
| compiler-analyzer (DP) | parse/canonicalize/score (RFC-002) | gRPC from SDK/CI; NATS jobs | blob(in-tenant), hash-cache Redis | job queue depth | analyses/min, cache hit, parser version counts | unknown IR → MEASUREMENT_ONLY; sandboxed (I.7) |
| runtime-analyzer (DP) | serving/NCCL/trace summarization | scrape + hooks | blob traces | scrape targets | ttft/tpot ingest rate | missing hooks → counter-only inference |
| control-plane-api | REST/gRPC edge, authn, rate-limit | public | none | RPS | 429 rate, authz denials | breaker to each domain svc; serves cached reads on partial outage |
| kernel-registry | kernel/toolchain records | gRPC | PG (kernels, toolchains) | RPS | registry size | outbox to Kafka |
| regression-svc | CRI compute, CI gate, reference DB | gRPC | PG + CH reads | gate QPS | CRI compute time, FP rate rolling | bench-runner quota guard |
| bench/verify-runner | isolated benchmark exec | jobs | ephemeral | queue | run variance | pinned env fingerprint; refuses noisy node (variance guard) |
| recommender | patterns → recs, ranking | gRPC | PG recs | events lag | recs/day, acceptance rate | prior-missing → wide-interval gains |
| graph-svc | ONLY writer to graph; named queries | gRPC | PG+AGE, pgvector | consumer lag | query p95 per named query | fallback to materialized views (D.10) |
| finance-svc | rates, cost slices, attribution | gRPC | Timescale | events lag | attribution completeness % | missing rates → slice parked+alert, never guessed |
| savings-svc | baselines, ledger, Shapley | gRPC | PG + Timescale | on-demand | dispute-replay exports | re-anchor single-flight lock (Redis) |
| twin-svc | analytical+residual sim | gRPC | model registry (blob) | sim queue | MAPE rolling, LOW_SUPPORT % | residual missing → analytical-only banner |
| policy-svc | OPA eval, toolchain approvals | gRPC | PG policies | RPS | decision latency, block counts | fail-closed for `block` policies, fail-open+audit for `warn` |
| audit-svc | hash chain writer/verifier | consumer + gRPC verify | PG + blob anchor | partition lag | chain gaps (must be 0) | gap ⇒ Sev-1 auto-page |
| agent-orchestrator | task lifecycle, MCP, budgets | NATS + gRPC | PG tasks, Redis scratch | task queue | tokens/task, judge-fail rate, schema-reject rate | model API down → park tasks; kill-switch flag |
| notify-svc | routes (email/slack/webhook) | consumer | PG subs | events | delivery failures | per-route breaker, dedup window |
| tenant-svc / auth-svc | provisioning, OIDC, keys, RLS GUC helpers | gRPC | PG | RPS | provision duration | offboarding = key-shred workflow (I.3) |
| ch-sink / ts-sink | Kafka→ClickHouse/Timescale loaders | consumers | CH/TS | lag | insert batch p95, late-row % | poison→DLQ; exactly-once via keeper offsets |

## G.2 Dependency graph (build-time enforced, RFC-012 O.6)
api → {registry, regression, recommender, graph, finance, savings, twin, policy, audit, tenant, agent-orch}; domain services → stores + bus only; DP services → gateway only. No service calls api tier. graph-svc is the sole graph writer; audit-svc the sole chain writer.

## G.3 Per-service SLOs (initial)
api availability 99.9%, read p95 300ms; ingest path (gateway→CH queryable) p95 <60s; kernel first-score freshness p95 <5min (L.1); named graph queries p95 <500ms; policy decision p95 <50ms (in deploy path); audit append p95 <200ms. Error budgets wired to release gating (RFC-011 J.6/J.11).
